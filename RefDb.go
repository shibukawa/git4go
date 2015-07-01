package git4go

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	GitPackedRefsFile        = "packed-refs"
	GitSymbolReference       = "ref: "
	PackPeelingNone     byte = 0
	PackPeelingStandard byte = 1
	PackPeelingFull     byte = 2
	PackRefHasPeel      byte = 1
	PackRefWasLoose     byte = 2
	PackRefCannotPeel   byte = 4
	PackRefShadowed     byte = 8
)

type PackRef struct {
	oid  *Oid
	peel *Oid
	flag byte
	name string
}

type PackRefSortedCache struct {
	lock           sync.RWMutex
	itemPathOffset int
	pool           bool
	items          []*PackRef
	cacheMap       map[string]*PackRef
	stamp          time.Time
	path           string
	peelingMode    byte
	notExist       bool
}

func (c *PackRefSortedCache) clear(lock bool) {
	if lock {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	c.cacheMap = make(map[string]*PackRef)
	c.items = []*PackRef{}
}

func (c *PackRefSortedCache) Upsert(key string) *PackRef {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.upsert(key)
}

func (c *PackRefSortedCache) upsert(key string) *PackRef {
	item, ok := c.cacheMap[key]
	if ok {
		return item
	}
	item = &PackRef{
		name: key,
	}
	c.cacheMap[key] = item
	c.items = append(c.items, item)
	return item
}

func (c *PackRefSortedCache) Lookup(key string) *PackRef {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.cacheMap[key]
}

type packRefs []*PackRef

func (p packRefs) Len() int {
	return len(p)
}
func (p packRefs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p packRefs) Less(i, j int) bool {
	return p[i].name < p[j].name
}

func (c *PackRefSortedCache) sort() {
	var items packRefs = c.items
	sort.Sort(items)
}

func (c *PackRefSortedCache) entries() []*PackRef {
	return c.items
}

func (c *PackRefSortedCache) remove(key string) {
	delete(c.cacheMap, key)
	for i, ref := range c.items {
		if ref.name == key {
			c.items = append(c.items[:i], c.items[i+1:]...)
			break
		}
	}
}

func (c *PackRefSortedCache) reloadIfChanged(lock bool) error {
	if lock {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	stat, err := os.Stat(c.path)
	if err != nil {
		c.notExist = true
		return nil
	}
	c.notExist = false
	if !c.stamp.Before(stat.ModTime()) {
		// not changed
		return nil
	}
	c.stamp = stat.ModTime()
	buffer, err := ioutil.ReadFile(c.path)

	if err != nil {
		c.clear(false)
		return err
	}

	scan := 0
	eof := len(buffer)

	c.peelingMode = PackPeelingNone
	if buffer[scan] == '#' {
		traitsHeader := []byte("# pack-refs with: ")
		if bytes.Equal(buffer[:len(traitsHeader)], traitsHeader) {
			scan += len(traitsHeader)
			eol := searchEndLine(buffer, scan)
			if eol == 0 {
				return errors.New("Corrupted packed references file")
			}
			line := buffer[scan:eol]
			if bytes.Index(line, []byte(" fully-peeled ")) != -1 {
				c.peelingMode = PackPeelingFull
			} else if bytes.Index(line, []byte(" peeled ")) != -1 {
				c.peelingMode = PackPeelingStandard
			}
			scan = eol + 1
		}
	}

	for scan < eof && buffer[scan] == '#' {
		eol := searchEndLine(buffer, scan)
		if eol == 0 {
			return errors.New("Corrupted packed references file")
		}
		scan = eol + 1
	}

	for scan < eof {
		oid, err := NewOid(string(buffer[scan : scan+GitOidHexSize]))
		if err != nil {
			return err
		}
		scan += GitOidHexSize
		if buffer[scan] != ' ' {
			return errors.New("Corrupted packed references file")
		}
		eol := searchEndLine(buffer, scan+1)
		if eol == 0 {
			return errors.New("Corrupted packed references file")
		}
		var line []byte
		if buffer[eol-1] == '\r' {
			line = buffer[scan+1 : eol-1]
		} else {
			line = buffer[scan+1 : eol]
		}
		ref := c.upsert(string(line))
		scan = eol + 1
		ref.oid = oid
		if scan < len(buffer) && buffer[scan] == '^' {
			peel, err := NewOid(string(buffer[scan+1 : scan+1+GitOidHexSize]))
			if err != nil {
				return err
			}
			scan += GitOidHexSize + 1
			if scan < eof {
				eol := searchEndLine(buffer, scan)
				if eol == 0 {
					return errors.New("Corrupted packed references file")
				}
				scan += GitOidHexSize + 1
			}
			ref.peel = peel
			ref.flag |= PackRefHasPeel
		} else if c.peelingMode == PackPeelingFull ||
			(c.peelingMode == PackPeelingStandard && strings.HasPrefix(ref.name, GitRefsTagsDir)) {
			ref.flag |= PackRefCannotPeel
		}
	}
	return nil
}

type RefDb struct {
	ignoreCase        bool
	precomposeUnicode bool
	repo              *Repository
	path              string
	cache             *PackRefSortedCache
}

func (r *Repository) NewRefDb() *RefDb {
	if r.refDb != nil {
		return r.refDb
	}

	if r.pathRepository == "" {
		return nil
	}

	config := r.Config()
	ignoreCase, _ := config.LookupBool("core.ignorecase")
	precomposeUnicode, _ := config.LookupBool("core.precomposeunicode")

	r.refDb = &RefDb{
		ignoreCase:        ignoreCase,
		precomposeUnicode: precomposeUnicode,
		repo:              r,
	}

	if r.namespace != "" {
		buffer := bytes.NewBufferString(r.pathRepository)
		for _, namespace := range strings.Split(r.namespace, "/") {
			buffer.WriteString("refs/namespaces/")
			buffer.WriteString(namespace)
			buffer.WriteByte('/')
		}
		buffer.WriteString("refs")
		r.refDb.path = buffer.String()
	} else {
		r.refDb.path = r.pathRepository
	}
	r.refDb.cache = &PackRefSortedCache{
		cacheMap: make(map[string]*PackRef),
		path:     filepath.Join(r.refDb.path, GitPackedRefsFile),
		stamp:    time.Unix(0, 0),
	}
	r.refDb.cache.reloadIfChanged(true)

	return r.refDb
}

func searchEndLine(buffer []byte, start int) int {
	eof := len(buffer)
	for i := start; i < eof; i++ {
		if buffer[i] == '\n' {
			return i
		}
	}
	return 0
}

func (r *RefDb) Lookup(name string) (*Reference, error) {
	refFile, err := ioutil.ReadFile(filepath.Join(r.path, name))
	if err == nil {
		refString := string(refFile)
		if strings.HasPrefix(refString, GitSymbolReference) {
			ref := &Reference{
				refType:        ReferenceSymbolic,
				targetSymbolic: strings.TrimSpace(refString[len(GitSymbolReference):]),
				repo:           r.repo,
			}
			return ref, nil
		} else {
			oid, err := NewOid(strings.TrimSpace(refString))
			if err != nil {
				return nil, err
			}
			ref := &Reference{
				refType:   ReferenceOid,
				targetOid: oid,
				repo:      r.repo,
				name:      name,
			}
			return ref, nil
		}
	} else {
		item := r.cache.Lookup(name)
		if item == nil {
			return nil, errors.New("not found")
		}
		ref := &Reference{
			refType:   ReferenceOid,
			targetOid: item.oid,
			repo:      r.repo,
			name:      name,
		}
		return ref, nil
	}
}

func (r *RefDb) GetPackedReferences() ([]*Reference, error) {
	r.cache.lock.Lock()
	defer r.cache.lock.Unlock()

	err := r.cache.reloadIfChanged(false)
	if err != nil {
		return nil, err
	}
	if r.cache.notExist {
		return []*Reference{}, nil
	}
	var result []*Reference
	for _, item := range r.cache.items {
		ref := &Reference{
			refType:   ReferenceOid,
			targetOid: item.oid,
			repo:      r.repo,
			name:      item.name,
		}
		result = append(result, ref)
	}
	return result, nil
}
