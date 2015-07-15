package git4go

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type TreeCache struct {
	children []*TreeCache

	entryCount int
	oid        *Oid
	name       string
}

func (v *TreeCache) write(buffer *bytes.Buffer) {
	buffer.WriteString(v.name)
	buffer.WriteByte(0)
	fmt.Fprintf(buffer, "%d %d\n", v.entryCount, len(v.children))
	if v.entryCount != -1 {
		buffer.Write(v.oid[:])
	}
	for _, child := range v.children {
		child.write(buffer)
	}
}

func (v *TreeCache) get(path string) *TreeCache {
	current := v
	for _, pathFragment := range strings.Split(path, "/") {
		if pathFragment == "" {
			continue
		}
		for _, child := range current.children {
			if child.name == pathFragment {
				current = child
				continue
			}
		}
		return nil
	}
	return current
}

func (v *TreeCache) invalidatePath(path string) {
	v.entryCount--
	current := v
	for _, pathFragment := range strings.Split(path, "/") {
		if pathFragment == "" {
			continue
		}
		for _, child := range current.children {
			if child.name == pathFragment {
				current = child
				current.entryCount--
				continue
			}
		}
		return /* we don't have that tree */
	}
}

func readTreeInternal(buffer []byte, offset, bufferEnd int) (*TreeCache, int, error) {
	nameEnd := findChar(buffer, 0, offset, bufferEnd)
	if nameEnd == -1 || bufferEnd-nameEnd < 8 {
		return nil, offset, errors.New("Corrupted TREE extension in index")
	}
	name := string(buffer[offset:nameEnd])
	offset = nameEnd + 1
	entryCount, newOffset := strtol32(buffer, offset, bufferEnd, 10)
	childCount, newOffset := strtol32(buffer, newOffset, bufferEnd, 10)
	if entryCount == -1 || childCount == -1 {
		return nil, offset, errors.New("Corrupted TREE extension in index")
	}

	cache := &TreeCache{
		name:       name,
		children:   make([]*TreeCache, childCount),
		entryCount: int(entryCount),
	}
	offset = newOffset
	if entryCount > 0 {
		if offset+GitOidRawSize > bufferEnd {
			return nil, offset, errors.New("Corrupted TREE extension in index")
		}
		cache.oid = NewOidFromBytes(buffer[offset : offset+GitOidRawSize])
		offset += GitOidRawSize
	}
	for i := 0; i < int(childCount); i++ {
		child, newOffset, err := readTreeInternal(buffer, offset, bufferEnd)
		if err != nil {
			return nil, offset, err
		}
		offset = newOffset
		cache.children[i] = child
	}
	return nil, offset, errors.New("Corrupted TREE extension in index")
}

func readTreeCache(buffer []byte, offset, extensionSize int) (*TreeCache, error) {
	tree, newOffset, err := readTreeInternal(buffer, offset, offset+extensionSize)
	if err != nil {
		return nil, err
	}
	if newOffset < offset+extensionSize {
		return tree, errors.New("Corrupted TREE extension in index (unexpected trailing data)")
	}
	return tree, nil
}

func readTreeCacheFromTreeRecursive(tree *Tree, cache *TreeCache) error {
	cache.oid = tree.Id()
	treeCount := 0
	for _, entry := range tree.Entries {
		if entry.Filemode == FilemodeTree {
			treeCount++
		}
	}
	cache.children = make([]*TreeCache, treeCount)
	entryCount := 0
	for i, entry := range tree.Entries {
		if entry.Filemode != FilemodeTree {
			entryCount++
			continue
		}
		childCache := &TreeCache{
			name: entry.Name,
		}
		childTree, err := tree.Owner().LookupTree(entry.Id)
		if err != nil {
			return err
		}
		err = readTreeCacheFromTreeRecursive(childTree, childCache)
		if err != nil {
			return err
		}
		cache.entryCount += childCache.entryCount
		cache.children[i] = childCache
	}
	return nil
}

func createTreeCacheFromTree(tree *Tree) error {
	cache := &TreeCache{}
	return readTreeCacheFromTreeRecursive(tree, cache)
}
