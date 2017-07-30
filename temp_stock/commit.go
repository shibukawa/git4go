package git4go

func (r *Repository) LookupCommit(id *Oid) (*Commit, error) {
	object, err := objectLookupPrefix(r, id, GIT_OID_HEXSZ, ObjectCommit)
	if err != nil {
		return nil, err
	}
	return object.(*Commit), nil
}

func (r *Repository) LookupCommitPrefix(id *Oid, length int) (*Commit, error) {
	object, err := objectLookupPrefix(r, id, length, ObjectCommit)
	if err != nil {
		return nil, err
	}
	return object.(*Commit), nil
}

type Commit struct {
	GitObject
}

func (c *Commit) Type() ObjectType {
	return ObjectCommit
}

func (c *Commit) Message() string {
	return ""
}

func (c *Commit) Summary() string {
	return ""
}

func (c *Commit) Tree() (*Tree, error) {
	return nil, nil
}

func (c *Commit) TreeId() *Oid {
	return nil
}

func (c *Commit) Author() *Signature {
	return nil
}

func (c *Commit) Committer() *Signature {
	return nil
}

func (c *Commit) Parent(n uint) *Commit {
	return nil
}

func (c *Commit) ParentId(n uint) *Oid {
	return nil
}

func (c *Commit) ParentCount() uint {
	return 0
}

/*func (c *Commit) Amend(refname string, author, committer *Signature, message string, tree *Tree) (*Oid, error) {
	return nil, nil
}*/
