package git4go

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
)

type SortType uint

const (
	SortNone        SortType = 1 << iota
	SortTopological SortType = 1 << iota
	SortTime        SortType = 1 << iota
	SortReverse     SortType = 1 << iota
)

func (v *Repository) Walk() (*RevWalk, error) {
	odb, err := v.Odb()
	if err != nil {
		return nil, err
	}
	revWalk := &RevWalk{
		repo:        v,
		odb:         odb,
		commits:     make(map[[20]byte]*commitListNode),
		getNext:     revWalkNextUnsorted,
		enqueue:     revWalkEnqueueUnsorted,
		walking:     false,
		firstParent: false,
		didHide:     false,
		didPush:     false,
	}
	return revWalk, nil
}

type getNextFunc func(revWalk *RevWalk) (*commitListNode, error)
type enqueueFunc func(revWalk *RevWalk, commit *commitListNode) error
type hideCbFunc func(oid *Oid) bool

type RevWalk struct {
	repo             *Repository
	odb              *Odb
	commits          map[[20]byte]*commitListNode
	topologyIterator commitListNodes
	randIterator     commitListNodes
	reverseIterator  commitListNodes
	timeIterator     commitListNodes
	userInput        commitListNodes

	getNext getNextFunc
	enqueue enqueueFunc
	//hideCb hideCbFunc

	walking     bool
	firstParent bool
	didHide     bool
	didPush     bool
	sorting     SortType
}

func (v *RevWalk) Reset() {
	for _, commit := range v.commits {
		commit.seen = false
		commit.inDegree = 0
		commit.topologyDelay = false
		commit.uninteresting = false
		commit.flags = 0
	}
	v.timeIterator = []*commitListNode{}
	v.randIterator = []*commitListNode{}
	v.reverseIterator = []*commitListNode{}
	v.userInput = []*commitListNode{}
	v.firstParent = false
	v.walking = false
	v.didPush = false
}

func (v *RevWalk) Push(id *Oid) error {
	return v.pushCommit(id, false, false)
}

func (v *RevWalk) PushGlob(glob string) error {
	return v.pushGlob(glob, false)
}

func (v *RevWalk) PushRange(r string) error {
	return nil
}

func (v *RevWalk) PushRef(r string) error {
	return v.pushRef(r, false, false)
}

func (v *RevWalk) PushHead() (err error) {
	return v.pushRef(GitHeadFile, false, false)
}

func (v *RevWalk) Hide(id *Oid) error {
	return v.pushCommit(id, true, false)
}

func (v *RevWalk) HideGlob(glob string) error {
	return v.pushGlob(glob, true)
}

func (v *RevWalk) HideRange(r string) error {
	return nil
}

func (v *RevWalk) HideRef(r string) error {
	return v.pushRef(r, true, false)
}

func (v *RevWalk) HideHead() (err error) {
	return v.pushRef(GitHeadFile, true, false)
}

func (v *RevWalk) Next(id *Oid) error {
	if !v.walking {
		err := v.prepareWalk()
		if err != nil {
			return err
		}
	}
	commit, err := v.getNext(v)
	if IsErrorCode(err, ErrIterOver) {
		v.Reset()
		return err
	}
	if err != nil {
		return err
	}
	copy(id[:], commit.oid[:])
	return nil
}

type RevWalkIterator func(commit *Commit) bool

func (v *RevWalk) Iterate(fun RevWalkIterator) (err error) {
	oid := new(Oid)
	for {
		err = v.Next(oid)
		if IsErrorCode(err, ErrIterOver) {
			return nil
		}
		if err != nil {
			return err
		}

		commit, err := v.repo.LookupCommit(oid)
		if err != nil {
			return err
		}

		cont := fun(commit)
		if !cont {
			break
		}
	}

	return nil
}

func (v *RevWalk) Sorting(sm SortType) {
	if v.walking {
		v.Reset()
	}
	v.sorting = sm
	if sm&SortTime != 0 {
		v.getNext = revWalkNextTimeSort
		v.enqueue = revWalkEnqueueTimeSort
	} else {
		v.getNext = revWalkNextUnsorted
		v.enqueue = revWalkEnqueueUnsorted
	}
}

func (v *RevWalk) premarkUninteresting() error {
	var q commitListNodes
	for _, commit := range v.userInput {
		err := v.commitListParse(commit)
		if err != nil {
			return err
		}
		q = q.insertByTime(commit)
	}
	for q.interesting() {
		commit := q[0]
		q = q[1:]
		for _, parent := range commit.parents {
			err := v.commitListParse(parent)
			if err != nil {
				return err
			}
			if commit.uninteresting {
				parent.uninteresting = true
			}
			if q.contains(parent) {
				continue
			}
			q = q.insertByTime(parent)
		}
	}
	return nil
}

func (v *RevWalk) markUninteresting(commit *commitListNode) error {
	var pending commitListNodes
	for {
		commit.uninteresting = true

		err := v.commitListParse(commit)
		if err != nil {
			return err
		}
		for _, parent := range commit.parents {
			if !parent.uninteresting {
				pending = append(pending, parent)
			}
		}
		if len(pending) > 0 {
			commit = pending[len(pending)-1]
			pending = pending[:len(pending)-1]
		} else {
			break
		}
		if pending.interestingArr() {
			break
		}
	}
	return nil
}

func (v *RevWalk) commitListParse(commit *commitListNode) error {
	if commit.parsed {
		return nil
	}
	obj, err := v.odb.Read(commit.oid)
	if err != nil {
		return err
	}
	if obj.Type != ObjectCommit {
		return errors.New("Object is no commit object")
	}
	return v.commitQuickParse(commit, obj.Data)
}

func (v *RevWalk) commitQuickParse(commit *commitListNode, data []byte) error {
	offset := 5 + GitOidHexSize + 1
	for {
		var parentId *Oid
		parentId, offset = parseOidWithPrefix(data, offset, []byte("parent "))
		if parentId == nil {
			break
		}
		parent := v.commitLookup(parentId)
		commit.parents = append(commit.parents, parent)
	}

	// skip author section
	found := false
	for offset < len(data) {
		if data[offset] == '\n' {
			found = true
			offset++
			break
		}
		offset++
	}
	if !found {
		return errors.New("object is corrupted")
	}

	timeSectionOffset := -1
	timeSectionEnd := -1
	afterSpace := false
	timeSection := false
	for offset < len(data) {
		c := data[offset]
		if c == '\n' {
			break
		}
		if timeSection {
			timeSectionEnd = offset
			if !('0' <= c && c <= '9') {
				timeSection = false
			}
		} else if afterSpace && '0' <= c && c <= '9' {
			timeSectionOffset = offset
			timeSection = true
		} else {
			afterSpace = (c == ' ')
		}
		offset++
	}
	if timeSectionOffset == -1 {
		return errors.New("object is corrupted")
	}
	timeStamp, err := strconv.ParseUint(string(data[timeSectionOffset:timeSectionEnd]), 10, 64)
	if err != nil {
		return err
	}
	commit.time = timeStamp
	commit.parsed = true
	return nil
}

func (v *RevWalk) prepareWalk() error {
	if !v.didPush {
		return MakeGitError("iteration over", ErrIterOver)
	}
	if v.didHide {
		err := v.premarkUninteresting()
		if err != nil {
			return err
		}
	}
	for _, commit := range v.userInput {
		err := v.processCommit(commit, commit.uninteresting)
		if err != nil {
			return err
		}
	}
	if (v.sorting & SortTopological) == SortTopological {
		next, err := v.getNext(v)
		for err == nil {
			for _, parent := range next.parents {
				parent.inDegree++
			}
			v.topologyIterator = append(v.topologyIterator, next)
			next, err = v.getNext(v)
		}
		if !IsErrorCode(err, ErrIterOver) {
			return err
		}
		v.getNext = revWalkNextTopologySort
	}

	if (v.sorting & SortReverse) == SortReverse {
		next, err := v.getNext(v)
		for err == nil {
			v.reverseIterator = append(v.reverseIterator, next)
			next, err = v.getNext(v)
		}
		if !IsErrorCode(err, ErrIterOver) {
			return err
		}
		v.getNext = revWalkNextReverse
	}
	v.walking = true
	return nil
}

func revWalkNextTimeSort(walk *RevWalk) (*commitListNode, error) {
	for {
		if len(walk.timeIterator) == 0 {
			break
		}
		next := walk.timeIterator[0]
		walk.timeIterator = walk.timeIterator[1:]
		if !next.uninteresting {
			err := walk.processCommitParents(next)
			if err != nil {
				return nil, err
			}
			return next, nil
		}
	}
	return nil, MakeGitError("iteration over", ErrIterOver)
}

func revWalkEnqueueTimeSort(walk *RevWalk, commit *commitListNode) error {
	walk.timeIterator = walk.timeIterator.insertByTime(commit)
	return nil
}

func revWalkNextUnsorted(walk *RevWalk) (*commitListNode, error) {
	for {
		length := len(walk.randIterator)
		if length == 0 {
			break
		}
		next := walk.randIterator[length-1]
		walk.randIterator = walk.randIterator[:length-1]
		if !next.uninteresting {
			err := walk.processCommitParents(next)
			if err != nil {
				return nil, err
			}
			return next, nil
		}
	}
	return nil, MakeGitError("iteration over", ErrIterOver)
}

func revWalkEnqueueUnsorted(walk *RevWalk, commit *commitListNode) error {
	walk.randIterator = append(walk.randIterator, commit)
	return nil
}

func revWalkNextTopologySort(walk *RevWalk) (*commitListNode, error) {
	for {
		length := len(walk.topologyIterator)
		if length == 0 {
			return nil, MakeGitError("iteration over", ErrIterOver)
		}
		next := walk.topologyIterator[length-1]
		walk.topologyIterator = walk.topologyIterator[:length-1]

		if next.inDegree > 0 {
			next.topologyDelay = true
			continue
		}

		max := len(next.parents)
		if walk.firstParent && len(next.parents) > 0 {
			max = 1
		}
		for i := 0; i < max; i++ {
			parent := next.parents[i]
			parent.inDegree--
			if parent.inDegree == 0 && parent.topologyDelay {
				parent.topologyDelay = false
				walk.topologyIterator = append(walk.topologyIterator, parent)
			}
		}

		return next, nil
	}
}

func revWalkNextReverse(walk *RevWalk) (*commitListNode, error) {
	length := len(walk.reverseIterator)
	if length == 0 {
		return nil, MakeGitError("iteration over", ErrIterOver)
	}
	next := walk.reverseIterator[length-1]
	walk.reverseIterator = walk.reverseIterator[:length-1]
	return next, nil
}

func (v *RevWalk) processCommit(commit *commitListNode, hide bool) error {
	/*if !hide && v.hideCb != nil {
		hide = v.hideCb(commit.oid)
	}*/
	if hide {
		err := v.markUninteresting(commit)
		if err != nil {
			return err
		}
	}
	if commit.seen {
		return nil
	}
	commit.seen = true
	err := v.commitListParse(commit)
	if err != nil {
		return err
	}
	if !hide {
		v.enqueue(v, commit)
	}
	return nil
}

func (v *RevWalk) processCommitParents(commit *commitListNode) error {
	max := len(commit.parents)
	if v.firstParent && max > 0 {
		max = 1
	}
	for i := 0; i < max; i++ {
		err := v.processCommit(commit.parents[i], commit.uninteresting)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *RevWalk) pushRef(refName string, uninteresting, fromGlob bool) error {
	ref, err := v.repo.LookupReference(refName)
	if err != nil {
		return err
	}
	resolved, err := ref.Resolve()
	if err != nil {
		return err
	}
	return v.pushCommit(resolved.Target(), uninteresting, fromGlob)
}

func (v *RevWalk) pushGlob(glob string, hide bool) error {
	if !strings.HasPrefix(glob, GitRefsDir) {
		glob = filepath.Join(GitRefsDir, glob)
	}
	if !strings.ContainsAny(glob, "?*[") {
		glob = filepath.Join(glob, "/*")
	}
	v.repo.ForEachGlobReference(glob, func(ref *Reference) error {
		resolved, err := ref.Resolve()
		if err != nil {
			return err
		}
		return v.pushCommit(resolved.Target(), hide, true)
	})
	return nil
}

func (v *RevWalk) pushCommit(oid *Oid, uninteresting, fromGlob bool) error {
	oobj, err := v.repo.Lookup(oid)
	if err != nil {
		return err
	}
	obj, err := oobj.Peel(ObjectCommit)
	if err != nil {
		if fromGlob {
			return nil
		}
		return err
	}
	commit := v.commitLookup(obj.Id())
	if commit.uninteresting {
		return nil
	}
	if uninteresting {
		v.didHide = true
	} else {
		v.didPush = true
	}
	commit.uninteresting = uninteresting
	v.userInput = append(v.userInput, commit)
	return nil
}

func (v *RevWalk) commitLookup(oid *Oid) *commitListNode {
	commit, ok := v.commits[*oid]
	if !ok {
		commit = &commitListNode{
			oid: oid,
		}
		v.commits[*oid] = commit
	}
	return commit
}
