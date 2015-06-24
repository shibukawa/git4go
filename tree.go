package git4go

import (
	"errors"
	"strconv"
)

type Filemode int

const (
	FilemodeTree           Filemode = 0040000
	FilemodeBlob           Filemode = 0100644
	FilemodeBlobExecutable Filemode = 0100755
	FilemodeLink           Filemode = 0120000
	FilemodeCommit         Filemode = 0160000
)

func (r *Repository) LookupTree(oid *Oid) (*Tree, error) {
	obj, err := objectLookupPrefix(r, oid, GitOidHexSize, ObjectTree)
	if obj != nil {
		return obj.(*Tree), err
	}
	return nil, err
}

func (r *Repository) LookupPrefixTree(oid *Oid, length int) (*Tree, error) {
	obj, err := objectLookupPrefix(r, oid, length, ObjectTree)
	if obj != nil {
		return obj.(*Tree), err
	}
	return nil, err
}

type Tree struct {
	gitObject
	Entries []*TreeEntry
}

func (t *Tree) EntryByName(filename string) *TreeEntry {
	for _, entry := range t.Entries {
		if entry.Name == filename {
			return entry
		}
	}
	return nil
}

func (t *Tree) EntryByPath(path string) (*TreeEntry, error) {
	return nil, nil
}

func (t *Tree) EntryByIndex(index int) *TreeEntry {
	if index < len(t.Entries) {
		return t.Entries[index]
	}
	return nil
}

func (t *Tree) EntryCount() uint64 {
	return uint64(len(t.Entries))
}

func (t *Tree) Type() ObjectType {
	return ObjectTree
}

func newTree(repo *Repository, oid *Oid, contents []byte) (*Tree, error) {
	var entries []*TreeEntry
	rawOffset := 0
	var name string
	for rawOffset < len(contents) {
		var attr int64
		for offset := rawOffset; offset < len(contents); offset++ {
			if contents[offset] == ' ' {
				attr, _ = strconv.ParseInt(string(contents[rawOffset:offset]), 8, 64)
				rawOffset = offset + 1
				break
			}
		}
		if attr == 0 {
			return nil, errors.New("Tree parse error: attribute")
		}
		for offset := rawOffset; offset < len(contents); offset++ {
			if contents[offset] == 0 {
				name = string(contents[rawOffset:offset])
				rawOffset = offset + 1
				break
			}
		}
		if name == "" {
			return nil, errors.New("Tree parse error: name")
		}
		oid := NewOidFromBytes(contents[rawOffset:])
		rawOffset += GitOidRawSize

		entry := &TreeEntry{
			Name:     name,
			Id:       oid,
			Type:     attr2oType(attr),
			Filemode: attr2Filemode(attr),
		}
		entries = append(entries, entry)
	}

	return &Tree{
		gitObject: gitObject{
			repo: repo,
			oid:  oid,
		},
		Entries: entries,
	}, nil
}

type TreeEntry struct {
	Name     string
	Id       *Oid
	Type     ObjectType
	Filemode Filemode
}

func attr2oType(attr int64) ObjectType {
	if (attr & 0170000 /* file type mask */) == int64(FilemodeCommit) {
		return ObjectCommit
	}
	if (attr & 0170000 /* file type mask */) == int64(FilemodeTree) {
		return ObjectTree
	}
	return ObjectBlob
}

func attr2Filemode(attr int64) Filemode {
	if (attr & 0170000 /* file type mask */) == int64(FilemodeTree) {
		return FilemodeTree
	}
	if (attr & 0111) != 0 {
		return FilemodeBlobExecutable
	}
	if (attr & 0170000 /* file type mask */) == int64(FilemodeCommit) {
		return FilemodeCommit
	}
	if (attr & 0170000 /* file type mask */) == int64(FilemodeLink) {
		return FilemodeLink
	}
	return FilemodeBlob
}
