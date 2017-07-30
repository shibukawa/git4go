package git4go

func (r *Repository) LookupTree(id *Oid) (*Commit, error) {
	return objectLookupPrefix(r, id, GIT_OID_HEXSZ, ObjectTree)
}

func (r *Repository) LookupTreePrefix(id *Oid, length int) (*Commit, error) {
	return objectLookupPrefix(r, id, length, ObjectTree)
}

type Tree struct {
	GitObject
}
