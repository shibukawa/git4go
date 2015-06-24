package git4go

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	GIT_PACKEDREFS_FILE = "packed-refs"
	GIT_SYMREF          = "ref: "
)

type RefDb struct {
	ignoreCase        bool
	precomposeUnicode bool
	repo              *Repository
	path              string
}

func (r *Repository) NewRefDb() *RefDb {
	if r.refDb != nil {
		return r.refDb
	}

	if r.pathRepository == "" {
		return nil
	}

	config := r.Config()
	ignoreCase, _ := config.LookupBool("core.ignorecase")
	precomposeUnicode, _ := config.LookupBool("core.precomposeunicode")

	r.refDb = &RefDb{
		ignoreCase:        ignoreCase,
		precomposeUnicode: precomposeUnicode,
		repo:              r,
	}

	if r.namespace != "" {
		buffer := bytes.NewBufferString(r.pathRepository)
		for _, namespace := range strings.Split(r.namespace, "/") {
			buffer.WriteString("refs/namespaces/")
			buffer.WriteString(namespace)
			buffer.WriteByte('/')
		}
		buffer.WriteString("refs")
		r.refDb.path = buffer.String()
	} else {
		r.refDb.path = r.pathRepository
	}

	//packedRefPath := filepath.Join(r.refDb.path, GIT_PACKEDREFS_FILE)

	return r.refDb
}

func (r *RefDb) Lookup(name string) (*Reference, error) {
	refFile, err := ioutil.ReadFile(filepath.Join(r.path, name))
	if err != nil {
		return nil, err
	}
	refString := string(refFile)
	if strings.HasPrefix(refString, GIT_SYMREF) {
		ref := &Reference{
			refType:        ReferenceSymbolic,
			targetSymbolic: strings.TrimSpace(refString[len(GIT_SYMREF):]),
			repo:           r.repo,
		}
		return ref, nil
	} else {
		oid, err := NewOid(strings.TrimSpace(refString))
		if err != nil {
			return nil, err
		}
		ref := &Reference{
			refType:   ReferenceOid,
			targetOid: oid,
			repo:      r.repo,
		}
		return ref, nil
	}
}
