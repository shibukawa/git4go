package git4go

import (
	"bytes"
)

func (r *Repository) LookupCommit(oid *Oid) (*Commit, error) {
	obj, err := objectLookupPrefix(r, oid, GitOidHexSize, ObjectCommit)
	if obj != nil {
		return obj.(*Commit), err
	}
	return nil, err
}

func (r *Repository) LookupPrefixCommit(oid *Oid, length int) (*Commit, error) {
	obj, err := objectLookupPrefix(r, oid, length, ObjectCommit)
	if obj != nil {
		return obj.(*Commit), err
	}
	return nil, err
}

type Commit struct {
	gitObject
	message   string
	summary   string
	treeId    *Oid
	author    *Signature
	committer *Signature
	Parents   []*Oid
}

func (t *Commit) Type() ObjectType {
	return ObjectCommit
}

func (c *Commit) Peel(targetType ObjectType) (Object, error) {
	return peel(c, targetType)
}

func (c *Commit) Message() string {
	return c.message
}

func (c *Commit) Summary() string {
	return c.summary
}

func (c Commit) Tree() (*Tree, error) {
	return c.repo.LookupTree(c.treeId)
}

func (c Commit) TreeId() *Oid {
	return c.treeId
}

func (c Commit) Author() *Signature {
	return c.author
}

func (c Commit) Committer() *Signature {
	return c.committer
}

func (c *Commit) Parent(n int) *Commit {
	return nil
}

func (c *Commit) ParentId(n int) *Oid {
	if n < len(c.Parents) {
		return c.Parents[n]
	}
	return nil
}

func (c *Commit) ParentCount() int {
	return len(c.Parents)
}

func (c *Commit) Amend(refname string, author, committer *Signature, message string, tree *Tree) (*Oid, error) {
	return nil, nil
}

func newCommit(repo *Repository, oid *Oid, contents []byte) (*Commit, error) {
	offset := 0
	var tree *Oid
	var parents []*Oid
	tree, offset = parseOidWithPrefix(contents, offset, []byte("tree "))
	for {
		var parent *Oid
		parent, offset = parseOidWithPrefix(contents, offset, []byte("parent "))
		if parent == nil {
			break
		}
		parents = append(parents, parent)
	}
	author, offset, err := parseSignature(contents, offset, []byte("author "))
	if err != nil {
		return nil, err
	}
	committer, offset, err := parseSignature(contents, offset, []byte("committer "))
	if err != nil {
		return nil, err
	}
	for offset < len(contents) {
		if contents[offset-1] == '\n' && contents[offset] == '\n' {
			break
		}
		eol := offset
		for eol < len(contents) && contents[eol] != '\n' {
			eol++
		}
		if bytes.Compare([]byte("encoding "), contents[offset:offset+len("encoding ")]) == 0 {
			offset += len("encoding ")
			// messageEncoding := contents[offset+len("encoding "):eol]
		}
		if eol < len(contents) && contents[eol] == '\n' {
			eol++
		}
		offset = eol
	}
	// rawHeader := contents[:offset]
	return &Commit{
		message:   string(contents[offset:]),
		treeId:    tree,
		author:    author,
		committer: committer,
		gitObject: gitObject{
			repo: repo,
			oid:  oid,
		},
	}, nil
}
