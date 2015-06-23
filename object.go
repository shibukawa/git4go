package git4go

import (
	"crypto/sha1"
	"errors"
	"fmt"
)

type ObjectType int

const (
	ObjectAny      ObjectType = -2
	ObjectBad      ObjectType = -2
	ObjectCommit   ObjectType = 1
	ObjectTree     ObjectType = 2
	ObjectBlob     ObjectType = 3
	ObjectTag      ObjectType = 4
	ObjectOfsDelta ObjectType = 6
	ObjectRefDelta ObjectType = 7
)

var typeString2Type map[string]ObjectType = map[string]ObjectType{
	"blob":   ObjectBlob,
	"commit": ObjectCommit,
	"tree":   ObjectTree,
	"tag":    ObjectTag,
}

var type2TypeString map[ObjectType]string = map[ObjectType]string{
	ObjectBlob:     "blob",
	ObjectCommit:   "commit",
	ObjectTree:     "tree",
	ObjectTag:      "tag",
	ObjectOfsDelta: "ofs-delta",
	ObjectRefDelta: "ref-delta",
}

func TypeString2Type(typeString string) ObjectType {
	objType, ok := typeString2Type[typeString]
	if !ok {
		return ObjectBad
	}
	return objType
}

func (o ObjectType) String() string {
	typeString, ok := type2TypeString[o]
	if !ok {
		return ""
	}
	return typeString
}

func hash(data []byte, objType ObjectType) (*Oid, error) {
	h := sha1.New()
	fmt.Fprintf(h, "%s %d\x00", objType.String(), len(data))
	h.Write(data)
	sha1Hash := h.Sum(nil)
	oid := new(Oid)
	copy(oid[:], sha1Hash[:])
	return oid, nil
}

type Object interface {
	Id() *Oid
	Type() ObjectType
	Owner() *Repository
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

func (r *Repository) Lookup(oid *Oid) (Object, error) {
	return objectLookupPrefix(r, oid, GitOidHexSize, ObjectAny)
}

func (r *Repository) LookupPrefix(oid *Oid, length int) (Object, error) {
	return objectLookupPrefix(r, oid, length, ObjectAny)
}

func objectLookupPrefix(repo *Repository, oid *Oid, length int, selectType ObjectType) (Object, error) {
	if length < GitOidMinimumPrefixLength {
		return nil, errors.New("Ambiguous lookup - OID prefix is too short")
	}

	if length > GitOidHexSize {
		length = GitOidHexSize
	}

	var rawObj *OdbObject
	var resultOid *Oid
	odb, err := repo.Odb()
	if err != nil {
		return nil, err
	}
	if length == GitOidHexSize {
		rawObj, err = odb.Read(oid)
		resultOid = oid
	} else {
		resultOid, rawObj, err = odb.ReadPrefix(oid, length)
	}
	if err != nil {
		return nil, err
	}
	if selectType != ObjectAny && rawObj.Type != selectType {
		return nil, errors.New("The requested type does not match the type in ODB")
	}
	switch rawObj.Type {
	case ObjectBlob:
		return newBlob(repo, resultOid, rawObj.Data), nil
	}
	return nil, errors.New("Invalid type:" + selectType.String())
}
