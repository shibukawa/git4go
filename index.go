package git4go

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/shibukawa/bsearch"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type IndexMatchedPathCallback func(string, string) IndexMatchResult
type IndexCompareFunc func(string, string) bool
type IndexAddOpts uint
type IndexEntryFlag uint16
type IndexEntryExtendedFlag uint16
type IndexCapFlag int
type IndexMatchResult int
type IndexStage int

const (
	GitIndexFile     = "index"
	GitIndexFileMode = 0666

	IndexAddDefault              IndexAddOpts = 0
	IndexAddForce                IndexAddOpts = 1
	IndexAddDisablePathspecMatch IndexAddOpts = 2
	IndexAddCheckPathspec        IndexAddOpts = 4

	IndexEntryNameMask   IndexEntryFlag = 0x0fff
	IndexEntryStageMask  IndexEntryFlag = 0x3000
	IndexEntryStageShift int            = 12
	IndexEntryExtended   uint16         = 0x4000

	IndexEntryIntentToAdd     IndexEntryExtendedFlag = 1 << 13
	IndexEntrySkipWorkTree    IndexEntryExtendedFlag = 1 << 14
	IndexEntryExtended2       IndexEntryExtendedFlag = 1 << 15
	IndexEntryExtendedFlags   IndexEntryExtendedFlag = IndexEntryIntentToAdd + IndexEntrySkipWorkTree
	IndexEntryUpdate          IndexEntryExtendedFlag = 1 << 0
	IndexEntryRemove          IndexEntryExtendedFlag = 1 << 1
	IndexEntryUpToDate        IndexEntryExtendedFlag = 1 << 2
	IndexEntryAdded           IndexEntryExtendedFlag = 1 << 3
	IndexEntryHashed          IndexEntryExtendedFlag = 1 << 4
	IndexEntryUnHashed        IndexEntryExtendedFlag = 1 << 5
	IndexEntryWTRemove        IndexEntryExtendedFlag = 1 << 6 /**< remove in work directory */
	IndexEntryConflicted      IndexEntryExtendedFlag = 1 << 7
	IndexEntryUnpacked        IndexEntryExtendedFlag = 1 << 8
	IndexEntryNewSkipWorkTree IndexEntryExtendedFlag = 1 << 9

	IndexHeaderSize = 12
	IndexFooterSize = 20

	IndexVersionNumber    = 2
	IndexVersionNumberExt = 3

	IndexHeaderSig uint32 = 0x44495243

	IndexMinimumEntrySize = 62

	IndexCapIgnoreCase IndexCapFlag = 1
	IndexCapNoFilemode IndexCapFlag = 2
	IndexCapNoSimlinks IndexCapFlag = 4
	IndexCapFromOwner  IndexCapFlag = -1

	IndexApplyFile IndexMatchResult = 0
	IndexSkipFile  IndexMatchResult = 1
	IndexAbort     IndexMatchResult = -1

	StageAncestor IndexStage = 1
	StageOurs     IndexStage = 2
	StageTheirs   IndexStage = 3
)

var IndexExtTreeCacheSig []byte = []byte("TREE")
var IndexExtUnmergedSig []byte = []byte("REUC")
var IndexExtConflictNameSig []byte = []byte("NAME")

type Index struct {
	repo             *Repository
	filePath         string
	stamp            int64
	entries          []*IndexEntry
	entriesSorted    bool
	lock             sync.Mutex
	deleted          []*IndexEntry
	readers          int
	onDisk           bool
	ignoreCase       bool
	distrustFilemode bool
	noSymlinks       bool

	tree       *TreeCache
	names      []*IndexNameEntry
	reuc       []*IndexReucEntry
	reucSorted bool
}

type IndexEntry struct {
	Ctime         time.Time
	Mtime         time.Time
	Mode          Filemode
	Uid           uint32
	Gid           uint32
	Size          uint32
	Id            *Oid
	Path          string
	flags         uint16
	flagsExtended uint16
}

type IndexNameEntry struct {
	ancestor string
	ours     string
	theirs   string
}

type IndexReucEntry struct {
	mode [3]Filemode
	oid  [3]*Oid
	path string
}

func (r *Repository) Index() (*Index, error) {
	if r.index == nil {
		index, err := OpenIndex(filepath.Join(r.pathRepository, GitIndexFile))
		if err != nil {
			return nil, err
		}
		index.repo = r
		r.index = index
		err = index.SetCaps(IndexCapFromOwner)
		if err != nil {
			return nil, err
		}
	}
	return r.index, nil
}

func (r *Repository) SetIndex(index *Index) {
	if r.index != nil {
		r.index.repo = nil
	}
	index.repo = r
	r.index = index
}

// NewIndex allocates a new index. It won't be associated with any
// file on the filesystem or repository
func NewIndex() (*Index, error) {
	return OpenIndex("")
}

// OpenIndex creates a new index at the given path. If the file does
// not exist it will be created when Write() is called.
func OpenIndex(path string) (*Index, error) {
	index := &Index{
		filePath: path,
		entries:  make([]*IndexEntry, 0, 32),
		names:    make([]*IndexNameEntry, 0, 8),
		reuc:     make([]*IndexReucEntry, 0, 8),
		deleted:  make([]*IndexEntry, 0, 8),
	}
	if path != "" {
		err := index.Read(true)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return index, nil
}

// Path returns the index' path on disk or an empty string if it
// exists only in memory.
func (v *Index) Path() string {
	return v.filePath
}

func (v *Index) Read(force bool) error {
	if v.filePath == "" {
		return errors.New("Failed to read index: The index is in-memory only")
	}
	stat, err := os.Stat(v.filePath)
	if os.IsNotExist(err) {
		v.onDisk = false
		if force {
			return v.Clear()
		}
		return nil
	}
	v.onDisk = true
	stamp := stat.ModTime().Unix()
	if v.stamp >= stamp && !force {
		return nil
	}
	buffer, err := ioutil.ReadFile(v.filePath)
	if err != nil {
		return err
	}
	err = v.Clear()
	if err != nil {
		return err
	}
	// check size and read checksum(sha1)
	if len(buffer) < IndexHeaderSize+IndexFooterSize {
		return errors.New("Index.Read(): insufficient buffer space")
	}
	calculatedChecksum := calcHash(buffer[:len(buffer)-IndexFooterSize])
	expectedChecksum := NewOidFromBytes(buffer[len(buffer)-IndexFooterSize:])
	if !calculatedChecksum.Equal(expectedChecksum) {
		return errors.New("Index.Read(): calculated checksum does not match expected")
	}

	// read header
	signature := ntohlFromBytes(buffer, 0)
	if signature != IndexHeaderSig {
		return errors.New("Index.Read(): incorrect header signature")
	}
	version := ntohlFromBytes(buffer, 4)
	if version != IndexVersionNumber && version != IndexVersionNumberExt {
		return errors.New("Index.Read(): incorrect header version")
	}
	entryCount := int(ntohlFromBytes(buffer, 8))
	// start reading entries
	v.lock.Lock()
	defer v.lock.Unlock()

	bound := len(buffer) - IndexFooterSize
	offset := IndexHeaderSize
	var i int
	for i = 0; i < entryCount && offset < bound; i++ {
		var entry *IndexEntry
		offset, entry = readEntry(buffer, offset)
		if entry != nil {
			v.entries = append(v.entries, entry)
		}
	}
	if i != entryCount {
		return errors.New("Index.Read(): header entries changed while parsing")
	}
	for offset < bound {
		size := readExtension(v, buffer, offset)
		if size == 0 {
			return errors.New("Index.Read(): extension is truncated")
		}
		offset += size
	}
	if offset != bound {
		return errors.New("buffer size does not match index footer size")
	}
	v.entriesSorted = !v.ignoreCase
	if !v.entriesSorted {
		v.sortEntriesIfNeeded(v.ignoreCase, false)
		v.entriesSorted = true
	}
	v.stamp = stamp
	return nil
}

func (v *Index) Clear() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.tree = nil
	v.entries = make([]*IndexEntry, 0, 32)
	v.names = make([]*IndexNameEntry, 0, 8)
	v.reuc = make([]*IndexReucEntry, 0, 8)
	v.deleted = make([]*IndexEntry, 0, 8)
	v.stamp = 0
	return nil
}

// Add adds or replaces the given entry to the index, making a copy of
// the data
func (v *Index) Add(entry *IndexEntry) error {
	if !validFilemode(entry.Mode) {
		return errors.New("invalid filemode")
	}
	v.entries = append(v.entries, entry)
	v.entriesSorted = false
	v.tree.invalidatePath(entry.Path)
	return nil
}

func (v *Index) AddByPath(path string) error {
	entry := &IndexEntry{
		Path: path,
	}
	v.entries = append(v.entries, entry)
	v.entriesSorted = false
	err := conflictToReuc(v, path)
	if err != nil {
		return err
	}
	v.tree.invalidatePath(path)
	return nil
}

func conflictToReuc(v *Index, path string) error {
	conflict, err := v.GetConflict(path)
	if err != nil {
		return err
	}
	var ancestorMode Filemode
	var ancestorOid *Oid
	if conflict.Ancestor != nil {
		ancestorMode = conflict.Ancestor.Mode
		ancestorOid = conflict.Ancestor.Id
	}
	var ourMode Filemode
	var ourOid *Oid
	if conflict.Our != nil {
		ourMode = conflict.Our.Mode
		ourOid = conflict.Our.Id
	}
	var theirMode Filemode
	var theirOid *Oid
	if conflict.Their != nil {
		theirMode = conflict.Their.Mode
		theirOid = conflict.Their.Id
	}
	err = reucAdd(v, path, ancestorMode, ourMode, theirMode, ancestorOid, ourOid, theirOid)
	if err != nil {
		return err
	}
	return v.RemoveConflict(path)

}

func reucAdd(v *Index, path string, ancestorMode, ourMode, theirMode Filemode, ancestorOid, ourOid, theirOid *Oid) error {
	reuc := &IndexReucEntry{
		mode: [3]Filemode{ancestorMode, ourMode, theirMode},
		oid:  [3]*Oid{ancestorOid, ourOid, theirOid},
		path: path,
	}
	index := reucFind(v, path)
	if index != -1 {
		v.reuc[index] = reuc
	} else {
		v.reuc = append(v.reuc, reuc)
		v.sortReuc(v.ignoreCase)
	}
	return nil
}

func reucFind(v *Index, reucPath string) int {
	if v.ignoreCase {
		reucPath = strings.ToLower(reucPath)
		return sort.Search(len(v.reuc), func(i int) bool {
			return strings.ToLower(v.reuc[i].path) == reucPath
		})
	} else {
		return sort.Search(len(v.reuc), func(i int) bool {
			return v.reuc[i].path == reucPath
		})
	}
}

func (v *Index) SetCaps(caps IndexCapFlag) error {
	if caps == IndexCapFromOwner {
		if v.repo == nil {
			return errors.New("Cannot access repository to set index caps")
		}
		conf := v.repo.Config()
		v.ignoreCase, _ = conf.LookupBooleanWithDefaultValue("core.ignorecase")
		filemode, _ := conf.LookupBooleanWithDefaultValue("core.filemode")
		v.distrustFilemode = !filemode
		symlinks, _ := conf.LookupBooleanWithDefaultValue("core.symlinks")
		v.noSymlinks = !symlinks
	} else {
		v.ignoreCase = caps&IndexCapIgnoreCase != 0
		v.distrustFilemode = caps&IndexCapNoFilemode != 0
		v.noSymlinks = caps&IndexCapNoSimlinks != 0
	}
	v.entriesSorted = false
	v.setIgnoreCase(v.ignoreCase)
	return nil
}

func (v *Index) setIgnoreCase(ignoreCase bool) {
	v.sortEntriesIfNeeded(ignoreCase, true)
	v.sortReuc(ignoreCase)
}

func (v *Index) Caps() IndexCapFlag {
	var flag IndexCapFlag
	if v.ignoreCase {
		flag |= IndexCapIgnoreCase
	}
	if v.distrustFilemode {
		flag |= IndexCapNoFilemode
	}
	if v.noSymlinks {
		flag |= IndexCapNoSimlinks
	}
	return flag
}

// todo
func (v *Index) AddAll(pathSpecs []string, flags IndexAddOpts, callback IndexMatchedPathCallback) error {
	return errors.New("not implemented")
}

// todo
func (v *Index) UpdateAll(pathSpecs []string, callback IndexMatchedPathCallback) error {
	return errors.New("not implemented")
}

// todo
func (v *Index) RemoveAll(pathSpecs []string, callback IndexMatchedPathCallback) error {
	return errors.New("not implemented")
}

func (v *Index) RemoveByPath(path string) error {
	err := v.Remove(path, 0)
	if err != nil && IsErrorCode(err, ErrNotFound) {
		return err
	}
	err = conflictToReuc(v, path)
	if err != nil && IsErrorCode(err, ErrNotFound) {
		return err
	}
	return nil
}

func (v *Index) Remove(path string, stage IndexStage) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	pos := v.sortAndFindInEntries(path, stage, false)
	if pos == -1 {
		return MakeGitError(fmt.Sprintf("Index does not contain %s at stage %d", path, stage), ErrNotFound)
	}
	return v.removeEntry(pos)
}

// todo
func (v *Index) WriteTreeTo(repo *Repository) (*Oid, error) {
	return nil, nil
}

// ReadTree replaces the contents of the index with those of the given
// tree
func (v *Index) ReadTree(tree *Tree) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	var newEntries []*IndexEntry

	v.sortEntriesIfNeeded(v.ignoreCase, false)
	err := tree.WalkPost(func(root string, treeEntry *TreeEntry) int {
		if treeEntry.Type == ObjectTree {
			return 0
		}
		path := filepath.Join(root, treeEntry.Name)
		entry := &IndexEntry{
			Path: path,
			Mode: treeEntry.Filemode,
			Id:   treeEntry.Id,
		}
		pos := v.findInEntries(v.entries, path, 0, false)
		if pos != -1 {
			oldEntry := v.entries[pos]
			if oldEntry.Mode == entry.Mode && oldEntry.Id.Equal(entry.Id) {
				entry.Ctime = oldEntry.Ctime
				entry.Mtime = oldEntry.Mtime
				entry.Uid = oldEntry.Uid
				entry.Gid = oldEntry.Gid
				entry.Size = oldEntry.Size
				entry.flagsExtended = 0
			}
		}
		if len(path) < int(IndexEntryNameMask) {
			entry.flags = uint16(IndexEntryFlag(len(path)) & IndexEntryNameMask)
		} else {
			entry.flags = uint16(IndexEntryNameMask)
		}

		newEntries = append(newEntries, entry)
		return 0
	})
	if err != nil {
		return err
	}
	if v.ignoreCase {
		var entries indexEntriesCaseInSensitive = newEntries
		sort.Sort(entries)
	} else {
		var entries indexEntriesCaseSensitive = newEntries
		sort.Sort(entries)
	}
	v.entries = newEntries
	return nil
}

// todo
func (v *Index) WriteTree() (*Oid, error) {
	return v.WriteTreeTo(v.repo)
}

// todo
func (v *Index) Write() error {
	return nil
}

func (v *Index) EntryCount() uint {
	return uint(len(v.entries))
}

func (v Index) Owner() *Repository {
	return v.repo
}

func (v *Index) EntryByIndex(index uint) (*IndexEntry, error) {
	v.sortEntriesIfNeeded(v.ignoreCase, true)
	if int(index) < len(v.entries) {
		return v.entries[index], nil
	}
	return nil, errors.New("out of index")
}

func (v *Index) Find(path string) int {
	v.lock.Lock()
	defer v.lock.Unlock()

	var pos int
	if v.ignoreCase {
		path = strings.ToLower(path)
		pos = bsearch.Search(len(v.entries), func(i int) bsearch.CompareResult {
			pathInList := strings.ToLower(v.entries[i].Path)
			if path > pathInList {
				return bsearch.Smaller
			} else if path == pathInList {
				return bsearch.Equal
			} else {
				return bsearch.Bigger
			}
		})
		if pos == -1 {
			return -1
		}
		// search the head element that has the specified path
		for ; pos > 0; pos-- {
			if strings.ToLower(v.entries[pos-1].Path) != path {
				return pos
			}
		}
	} else {
		pos = bsearch.Search(len(v.reuc), func(i int) bsearch.CompareResult {
			pathInList := v.entries[i].Path
			if path > pathInList {
				return bsearch.Smaller
			} else if path == pathInList {
				return bsearch.Equal
			} else {
				return bsearch.Bigger
			}
		})
		if pos == -1 {
			return -1
		}
		// search the head element that has the specified path
		for ; pos > 0; pos-- {
			if v.entries[pos-1].Path != path {
				return pos
			}
		}
	}
	return 0
}

func (v *Index) findInEntries(entries []*IndexEntry, path string, stage IndexStage, ignoreCase bool) int {
	if v.ignoreCase {
		path = strings.ToLower(path)
		return bsearch.Search(len(entries), func(i int) bsearch.CompareResult {
			pathInList := strings.ToLower(entries[i].Path)
			if path > pathInList {
				return bsearch.Smaller
			} else if path == pathInList {
				return bsearch.CompareResult(entries[i].Stage() - stage)
			} else {
				return bsearch.Bigger
			}
		})
	} else {
		return bsearch.Search(len(v.reuc), func(i int) bsearch.CompareResult {
			pathInList := entries[i].Path
			if path > pathInList {
				return bsearch.Smaller
			} else if path == pathInList {
				return bsearch.CompareResult(entries[i].Stage() - stage)
			} else {
				return bsearch.Bigger
			}
		})
	}
}

func (v *Index) sortAndFindInEntries(path string, stage IndexStage, needLock bool) int {
	v.sortEntriesIfNeeded(v.ignoreCase, needLock)
	return v.findInEntries(v.entries, path, stage, v.ignoreCase)
}

func (v *Index) HasConflicts() bool {
	for _, entry := range v.entries {
		if entry.Stage() != 0 {
			return false
		}
	}
	return true
}

// FIXME: this might return an error
func (v *Index) CleanupConflicts() {
}

func (v *Index) AddConflict(ancestor *IndexEntry, our *IndexEntry, their *IndexEntry) error {
	return nil
}

type IndexConflict struct {
	Ancestor *IndexEntry
	Our      *IndexEntry
	Their    *IndexEntry
}

func (v *Index) GetConflict(path string) (IndexConflict, error) {
	index := v.Find(path)
	if index < 0 {
		return IndexConflict{}, MakeGitError("Index.GetConflict(): not found: "+path, ErrNotFound)
	}
	conflict, length := v.getConflictByIndex(index)
	if length < 0 {
		return IndexConflict{}, MakeGitError("Index.GetConflict(): not found: "+path, ErrNotFound)
	}
	return conflict, nil
}

func (v *Index) RemoveConflict(path string) error {
	return nil
}

func (v IndexEntry) Stage() IndexStage {
	return IndexStage((v.flags & uint16(IndexEntryStageMask)) >> uint16(IndexEntryStageShift))
}

func (v *IndexEntry) SetStage(flag IndexStage) {
	v.flags = (v.flags & (^uint16(IndexEntryStageMask))) | ((uint16(flag) & 0x03) << uint16(IndexEntryStageShift))
}

func (v IndexEntry) IsConflict() bool {
	return v.Stage() != 0
}

type IndexConflictIterator struct {
	index  *Index
	cursor int
}

func (v *IndexConflictIterator) Index() *Index {
	return v.index
}

func (v *Index) ConflictIterator() (*IndexConflictIterator, error) {
	iter := &IndexConflictIterator{
		index:  v,
		cursor: 0,
	}
	return iter, nil
}

func (v *IndexConflictIterator) Next() (IndexConflict, error) {
	for v.cursor < len(v.index.entries) {
		entry := v.index.entries[v.cursor]
		if entry.IsConflict() {
			conflict, length := v.index.getConflictByIndex(v.cursor)
			v.cursor += length
			return conflict, nil
		}
		v.cursor++
	}
	return IndexConflict{}, MakeGitError("IndexConflictIterator.Next(): iterator is over", ErrIterOver)
}

func (v *Index) getConflictByIndex(pos int) (IndexConflict, int) {
	var path string
	length := 0
	result := IndexConflict{}
	for _, entry := range v.entries[pos:] {
		if path != "" && entry.Path != path {
			break
		}
		path = entry.Path
		switch entry.Stage() {
		case StageAncestor:
			result.Ancestor = entry
			length++
		case StageOurs:
			result.Our = entry
			length++
		case StageTheirs:
			result.Their = entry
			length++
		}
	}
	return result, length
}

// index entries
type indexEntriesCaseSensitive []*IndexEntry

func (a indexEntriesCaseSensitive) Len() int {
	return len(a)
}
func (a indexEntriesCaseSensitive) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a indexEntriesCaseSensitive) Less(i, j int) bool {
	e1 := a[i]
	e2 := a[j]
	if e1.Path < e2.Path {
		return true
	} else if e1.Path == e2.Path {
		return e1.Stage() < e2.Stage()
	} else {
		return false
	}
}

type indexEntriesCaseInSensitive []*IndexEntry

func (a indexEntriesCaseInSensitive) Len() int {
	return len(a)
}
func (a indexEntriesCaseInSensitive) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a indexEntriesCaseInSensitive) Less(i, j int) bool {
	e1 := a[i]
	e2 := a[j]
	e1Path := strings.ToLower(e1.Path)
	e2Path := strings.ToLower(e2.Path)
	if e1Path < e2Path {
		return true
	} else if e1Path == e2Path {
		return e1.Stage() < e2.Stage()
	} else {
		return false
	}
}

// reuc entries
type reucEntriesCaseSensitive []*IndexReucEntry

func (a reucEntriesCaseSensitive) Len() int {
	return len(a)
}
func (a reucEntriesCaseSensitive) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a reucEntriesCaseSensitive) Less(i, j int) bool {
	return a[i].path < a[j].path
}

type reucEntriesCaseInSensitive []*IndexReucEntry

func (a reucEntriesCaseInSensitive) Len() int {
	return len(a)
}
func (a reucEntriesCaseInSensitive) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a reucEntriesCaseInSensitive) Less(i, j int) bool {
	return strings.ToLower(a[i].path) < strings.ToLower(a[j].path)
}

func (v *Index) sortEntriesIfNeeded(ignoreCase, lock bool) {
	if v.entriesSorted {
		return
	}
	if lock {
		v.lock.Lock()
		defer v.lock.Unlock()
	}
	if ignoreCase {
		var entries indexEntriesCaseInSensitive = v.entries
		sort.Sort(entries)
	} else {
		var entries indexEntriesCaseSensitive = v.entries
		sort.Sort(entries)
	}
}

func (v *Index) sortReuc(ignoreCase bool) {
	if ignoreCase {
		var reuc reucEntriesCaseInSensitive = v.reuc
		sort.Sort(reuc)
	} else {
		var reuc reucEntriesCaseSensitive = v.reuc
		sort.Sort(reuc)
	}
}

type indexNameEntries []*IndexNameEntry

func (a indexNameEntries) Len() int {
	return len(a)
}
func (a indexNameEntries) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a indexNameEntries) Less(i, j int) bool {
	e1 := a[i]
	e2 := a[j]
	if e1.ancestor < e2.ancestor {
		return true
	} else if e1.ancestor > e2.ancestor {
		return false
	}
	return e1.ours < e2.ours
}

func readEntry(buffer []byte, offset int) (int, *IndexEntry) {
	bound := len(buffer) - IndexFooterSize
	if offset+IndexMinimumEntrySize > bound {
		return offset, nil
	}
	entry := &IndexEntry{
		Ctime: time.Unix(int64(ntohlFromBytes(buffer, offset)), int64(ntohlFromBytes(buffer, offset+4))),
		Mtime: time.Unix(int64(ntohlFromBytes(buffer, offset+8)), int64(ntohlFromBytes(buffer, offset+12))),
		//Dev: ntohlFromBytes(buffer, offset+16),
		//Ino: ntohlFromBytes(buffer, offset+20),
		Mode:  Filemode(ntohlFromBytes(buffer, offset+24)),
		Uid:   ntohlFromBytes(buffer, offset+28),
		Gid:   ntohlFromBytes(buffer, offset+32),
		Size:  ntohlFromBytes(buffer, offset+36),
		Id:    NewOidFromBytes(buffer[offset+40 : offset+60]),
		flags: ntohsFromBytes(buffer, offset+60),
	}
	var pathStart int
	if entry.flags&IndexEntryExtended != 0 {
		entry.flagsExtended = ntohsFromBytes(buffer, offset+62)
		pathStart = offset + 64
	} else {
		pathStart = offset + 62
	}
	pathLength := int(entry.flags & uint16(IndexEntryNameMask))
	if pathLength == int(IndexEntryNameMask) {
		pathEnd := pathStart
		found := false
		for pathEnd < bound {
			if buffer[pathEnd] == 0 {
				found = true
				break
			}
		}
		if !found {
			return offset, nil
		}
		pathLength = pathEnd - pathStart
	}
	entry.Path = string(buffer[pathStart : pathStart+pathLength])
	offset = ((pathStart + pathLength + 8 - offset) & ^7) + offset
	return offset, entry
}

func readReuc(index *Index, buffer []byte, offset, size int) error {
	for size > 0 {
		pathEnd := findChar(buffer, 0, offset, offset+size)
		lost := &IndexReucEntry{
			path: string(buffer[offset:pathEnd]),
		}
		size -= (pathEnd - offset + 1)
		offset = pathEnd + 1

		for i := 0; i < 3; i++ {
			tmp, nextOffset := strtol32(buffer, offset, offset+size, 8)
			if tmp < 0 || tmp > 0xffffffff {
				return errors.New("reading reuc entry stage")
			}
			lost.mode[i] = Filemode(tmp)
			size -= nextOffset - offset
			offset = nextOffset
			if size < 0 {
				return errors.New("reading reuc entry stage")
			}
		}
		for i := 0; i < 3; i++ {
			if lost.mode[i] == 0 {
				continue
			}
			if size < 20 {
				return errors.New("reading reuc entry oid")
			}
			lost.oid[i] = NewOidFromBytes(buffer[offset : offset+GitOidRawSize])
			offset += 20
			size -= 20
		}
		index.reuc = append(index.reuc, lost)
	}
	index.reucSorted = true
	return nil
}

func readConflictNames(index *Index, buffer []byte, offset, size int) error {
	for size > 0 {
		ancestor, nextOffset := readString(buffer, offset, offset+size)
		if nextOffset < 0 {
			goto readError
		}
		ours, nextOffset := readString(buffer, nextOffset, offset+size)
		if nextOffset < 0 {
			goto readError
		}
		theirs, nextOffset := readString(buffer, nextOffset, offset+size)
		if nextOffset < 0 {
			goto readError
		}
		conflictName := &IndexNameEntry{
			ancestor: ancestor,
			ours:     ours,
			theirs:   theirs,
		}
		index.names = append(index.names, conflictName)
		size -= (nextOffset - offset)
		offset = nextOffset
	}
	return nil
readError:
	return errors.New("reading conflict name entries")
}

func readExtension(index *Index, buffer []byte, offset int) int {
	extensionSize := int(ntohlFromBytes(buffer, offset+4))
	totalSize := extensionSize + 8
	if offset+totalSize > len(buffer)-IndexFooterSize {
		return 0
	}
	c := buffer[offset]
	if 'A' <= c && c <= 'Z' {
		sig := buffer[offset : offset+4]
		if bytes.Equal(sig, IndexExtTreeCacheSig) {
			cache, _ := readTreeCache(buffer, offset+8, extensionSize)
			if cache == nil {
				return 0
			}
			index.tree = cache
		} else if bytes.Equal(sig, IndexExtUnmergedSig) {
			err := readReuc(index, buffer, offset+8, extensionSize)
			if err != nil {
				return 0
			}
		} else if bytes.Equal(sig, IndexExtConflictNameSig) {
			err := readConflictNames(index, buffer, offset+8, extensionSize)
			if err != nil {
				return 0
			}
		}
	} else {
		return 0
	}
	return totalSize
}

func (v *Index) removeEntry(pos int) error {
	entry := v.entries[pos]
	v.tree.invalidatePath(entry.Path)
	v.entries = append(v.entries[:pos], v.entries[pos+1:]...)
	if v.readers > 0 {
		v.deleted = append(v.deleted, entry)
	}
	return nil
}
