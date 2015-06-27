package git4go

import (
	"bytes"
	"errors"
	"strings"
)

func (r *Repository) LookupTag(oid *Oid) (*Tag, error) {
	obj, err := objectLookupPrefix(r, oid, GitOidHexSize, ObjectTag)
	if obj != nil {
		return obj.(*Tag), err
	}
	return nil, err
}

func (r *Repository) LookupPrefixTag(oid *Oid, length int) (*Tag, error) {
	obj, err := objectLookupPrefix(r, oid, length, ObjectTag)
	if obj != nil {
		return obj.(*Tag), err
	}
	return nil, err
}

func (r *Repository) ListTag() ([]string, error) {
	var tags []string
	err := r.ForEachReferenceName(func(path string) error {
		if strings.HasPrefix(path, GitTagsDir) {
			tags = append(tags, path[len(GitTagsDir)+1:])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

type Tag struct {
	gitObject
	targetType ObjectType
	targetId   *Oid
	tagger     *Signature
	message    string
	name       string
}

func (t *Tag) Type() ObjectType {
	return ObjectTag
}

func (t *Tag) Message() string {
	return t.message
}

func (t *Tag) Name() string {
	return t.name
}

func (t *Tag) Tagger() *Signature {
	return t.tagger
}

func (t *Tag) Target() Object {
	obj, _ := t.repo.Lookup(t.targetId)
	return obj
}

func (t *Tag) TargetId() *Oid {
	return t.targetId
}

func (t *Tag) TargetType() ObjectType {
	return t.targetType
}

func newTag(repo *Repository, oid *Oid, contents []byte) (*Tag, error) {
	targetId, offset := parseOidWithPrefix(contents, 0, []byte("object "))
	if targetId == nil {
		return nil, errors.New("Object field invalid")
	}
	if len(contents)-offset < 5 {
		return nil, errors.New("Object too short")
	}
	if !bytes.Equal(contents[offset:offset+5], []byte("type ")) {
		return nil, errors.New("Type field not found")
	}
	offset += 5
	targetType := ObjectBad
	for eol := offset; eol < len(contents); eol++ {
		if contents[eol] == '\n' {
			targetType = TypeString2Type(string(contents[offset:eol]))
			offset = eol + 1
			break
		}
	}
	if targetType == ObjectBad {
		return nil, errors.New("Invalid object type")
	}
	if len(contents)-offset < 4 {
		return nil, errors.New("Object too short")
	}
	if !bytes.Equal(contents[offset:offset+4], []byte("tag ")) {
		return nil, errors.New("Tag field not found")
	}
	offset += 4
	tagName := ""
	for eol := offset; eol < len(contents); eol++ {
		if contents[eol] == '\n' {
			tagName = string(contents[offset:eol])
			offset = eol + 1
			break
		}
	}

	var tagger *Signature
	var err error
	tagger, offset, err = parseSignature(contents, offset, []byte("tagger "))
	if err != nil {
		return nil, err
	}
	message := ""
	if offset < len(contents) {
		if contents[offset] != '\n' {
			return nil, errors.New("No new line before message")
		}
		message = string(contents[offset+1:])
	}

	return &Tag{
		gitObject: gitObject{
			repo: repo,
			oid:  oid,
		},
		name:       tagName,
		message:    message,
		tagger:     tagger,
		targetId:   targetId,
		targetType: targetType,
	}, nil
}
