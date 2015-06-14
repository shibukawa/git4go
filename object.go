package git4go

import (
//"errors"
)

type ObjectType int

const (
	ObjectAny    ObjectType = -2
	ObjectBad    ObjectType = -2
	ObjectCommit ObjectType = 1
	ObjectTree   ObjectType = 2
	ObjectBlob   ObjectType = 3
	ObjectTag    ObjectType = 4
)

/*
type Object interface {
	Id() *Oid
	Type() ObjectType
	Owner() *Repository
	Size() int
}

type GitObject struct {
	repo *Repository
	oid  *Oid
}

func (o *GitObject) Owner() *Repository {
	return o.repo
}

func (o *GitObject) Id() *Oid {
	return o.oid
}

func (r *Repository) Lookup(id *Oid) (Object, error) {
	return objectLookupPrefix(r, id, GIT_OID_HEXSZ, ObjectAny)
}

func (r *Repository) LookupPrefix(id *Oid, length int) (Object, error) {
	return objectLookupPrefix(r, id, length, ObjectAny)
}

func objectLookupPrefix(repo *Repository, oid *Oid, length int, selectType ObjectType) (Object, error) {
	if length < GIT_OID_MINPREFIXLEN {
		return errors.New("Ambiguous lookup - OID prefix is too short")
	}

	if length > GIT_OID_HEXSZ {
		length = GIT_OID_HEXSZ
	}

	if length == GIT_OID_HEXSZ {
		cachedObject := repo.cache.Get(oid)
		if cachedObject != nil {
			if selectType != ObjectAny && selectType != cachedObject.Type() {
				return nil, errors.New("The requested type does not match the type in ODB")
			}
			return cachedObject, nil
		} else {
			odb, _ := repo.Odb()
			return odb.read(oid)
		}
	} else {
		shortOid := new(Oid)
		copy(shortOid[:], oid[:(length+1)/2])
		if (length % 2) == 1 {
			shortOid[len/2] &= 0xF0
		}
		odb, _ := repo.Odb()
		return odb.readPrefix(oid, length)
	}
}
*/

var typeString2Type map[string]ObjectType = map[string]ObjectType{
	"blob":   ObjectBlob,
	"commit": ObjectCommit,
	"tree":   ObjectTree,
	"tag":    ObjectTag,
}

func TypeString2Type(typeString string) ObjectType {
	objType, ok := typeString2Type[typeString]
	if !ok {
		return ObjectBad
	}
	return objType
}
