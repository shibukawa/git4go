package git4go

func (r *Repository) LookupBlob(oid *Oid) (*Blob, error) {
	obj, err := objectLookupPrefix(r, oid, GitOidHexSize, ObjectBlob)
	return obj.(*Blob), err
}

func (r *Repository) LookupPrefixBlob(oid *Oid, length int) (*Blob, error) {
	obj, err := objectLookupPrefix(r, oid, length, ObjectBlob)
	return obj.(*Blob), err
}

func (repo *Repository) CreateBlobFromBuffer(data []byte) (*Oid, error) {
	odb, err := repo.Odb()
	if err != nil {
		return nil, err
	}
	return odb.Write(data, ObjectBlob)
}

type Blob struct {
	gitObject
	contents []byte
}

func (b *Blob) Type() ObjectType {
	return ObjectBlob
}

func (b *Blob) Peel(targetType ObjectType) (Object, error) {
	return peel(b, targetType)
}

func (b *Blob) Size() int64 {
	return int64(len(b.contents))
}

func (b *Blob) Contents() []byte {
	return b.contents
}

func newBlob(repo *Repository, oid *Oid, contents []byte) *Blob {
	return &Blob{
		contents: contents,
		gitObject: gitObject{
			repo: repo,
			oid:  oid,
		},
	}
}
