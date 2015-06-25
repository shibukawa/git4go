package git4go

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

func (r *Repository) TreeBuilder() (*TreeBuilder, error) {
	return &TreeBuilder{
		repo:    r,
		Entries: make(map[string]*TreeEntry),
	}, nil
}

type TreeBuilder struct {
	repo    *Repository
	Entries map[string]*TreeEntry
}

type TreeEntries []*TreeEntry

func (p TreeEntries) Len() int {
	return len(p)
}
func (p TreeEntries) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p TreeEntries) Less(i, j int) bool {
	return p[i].Name < p[j].Name
}

func (b *TreeBuilder) Insert(filename string, oid *Oid, filemode Filemode) error {
	if oid == nil {
		return errors.New("oid should not be nil")
	}
	entry := &TreeEntry{
		Name:     filename,
		Id:       oid,
		Filemode: filemode,
		Type:     attr2oType(int64(filemode)),
	}
	b.Entries[filename] = entry
	return nil
}

func (b *TreeBuilder) Remove(filename string) error {
	delete(b.Entries, filename)
	return nil
}

func (b *TreeBuilder) Write() (*Oid, error) {
	odb, err := b.repo.Odb()
	if err != nil {
		return nil, err
	}
	var entries TreeEntries
	for _, entry := range b.Entries {
		entries = append(entries, entry)
	}
	sort.Sort(entries)
	var buffer = bytes.NewBuffer(make([]byte, 0, len(entries)*72))
	for _, entry := range entries {
		fmt.Fprintf(buffer, "%o %s", int(entry.Filemode), entry.Name)
		buffer.WriteByte(0)
		buffer.Write(entry.Id[:])
	}
	return odb.Write(buffer.Bytes(), ObjectTree)
}
