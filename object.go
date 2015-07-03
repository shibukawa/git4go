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
	Peel(targetType ObjectType) (Object, error)
}

type gitObject struct {
	repo *Repository
	oid  *Oid
}

func (o *gitObject) Owner() *Repository {
	return o.repo
}

func (o *gitObject) Id() *Oid {
	return o.oid
}

func checkTypeCombination(sourceType, targetType ObjectType) bool {
	if sourceType == targetType {
		return true
	}
	switch sourceType {
	case ObjectBlob:
		return false
	case ObjectTree:
		return false
	case ObjectCommit:
		return (targetType == ObjectTree) || (targetType == ObjectAny)
	case ObjectTag:
		return true
	}
	return false
}

func peelError(oid *Oid, targetType ObjectType) error {
	msg := fmt.Sprintf("The git_object of id '%s' can not be successfully peeled into a %s.", oid, targetType)
	return errors.New(msg)
}

func dereferenceObject(object Object) Object {
	switch object.Type() {
	case ObjectCommit:
		tree, _ := object.(*Commit).Tree()
		return tree
	case ObjectTag:
		return object.(*Tag).Target()
	default:
		return nil
	}
}

func peel(source Object, targetType ObjectType) (Object, error) {
	if targetType != ObjectTag && targetType != ObjectCommit && targetType != ObjectTree && targetType != ObjectBlob && targetType != ObjectAny {
		return nil, errors.New("invalid type")
	}
	sourceType := source.Type()
	if !checkTypeCombination(sourceType, targetType) {
		peelError(source.Id(), targetType)
	}
	if source.Type() == targetType {
		return source, nil
	}
	for {
		peeled := dereferenceObject(source)
		if peeled == nil {
			break
		}
		if peeled.Type() == targetType || (targetType == ObjectAny && peeled.Type() != sourceType) {
			return peeled, nil
		}
		source = peeled
	}
	return nil, peelError(source.Id(), targetType)
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
	case ObjectTree:
		return newTree(repo, resultOid, rawObj.Data)
	case ObjectCommit:
		return newCommit(repo, resultOid, rawObj.Data)
	case ObjectTag:
		return newTag(repo, resultOid, rawObj.Data)
	}
	return nil, errors.New("Invalid type:" + selectType.String())
}
