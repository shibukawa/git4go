package git4go

import (
	"errors"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"os"
	"path/filepath"
	"strings"
)

type ReferenceType int

const (
	ReferenceOid        ReferenceType = 1
	ReferenceSymbolic   ReferenceType = 2
	DefaultNestingLevel               = 5
	MaxNestingLevel                   = 10
	GitRefNameMax                     = 1024
)

// Repository methods related to Reference

func (r *Repository) LookupReference(name string) (*Reference, error) {
	return referenceLookupResolved(r, name, 0)
}

func (r *Repository) Head() (*Reference, error) {
	head, err := r.LookupReference(GitHeadFile)
	if head.Type() == ReferenceOid {
		return head, err
	}
	return referenceLookupResolved(r, head.targetSymbolic, -1)
}

var dwimReferenceFormatter []string = []string{
	"%s",
	"refs/%s",
	"refs/tags/%s",
	"refs/heads/%s",
	"refs/remotes/%s",
	"refs/remotes/%s/HEAD",
}

func (r *Repository) DwimReference(name string) (*Reference, error) {
	if name == "" {
		name = GitHeadFile
	}
	for _, formatter := range dwimReferenceFormatter {
		refName := fmt.Sprintf(formatter, name)
		refName2, err := referenceNormalize(refName, false, true)
		if err != nil {
			return nil, err
		}
		ref, _ := referenceLookupResolved(r, refName2, -1)
		if ref != nil {
			return ref, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Could not use '%s' as valid reference name", name))
}

type ForEachReferenceNameCallback func(string) error

func (r *Repository) ForEachReferenceName(callback ForEachReferenceNameCallback) error {
	rootDir := filepath.Join(r.pathRepository, GitRefsDir)
	processed := make(map[string]bool)
	offset := len(r.pathRepository)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = path[offset:]
		processed[path] = true
		return callback(path)
	})
	if err != nil {
		return err
	}
	refDb := r.NewRefDb()
	refs, err := refDb.GetPackedReferences()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if !processed[ref.name] {
			err = callback(ref.name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type ForEachReferenceCallback func(*Reference) error

func (r *Repository) ForEachReference(callback ForEachReferenceCallback) error {
	rootDir := filepath.Join(r.pathRepository, GitRefsDir)
	processed := make(map[string]bool)
	offset := len(rootDir) - 4
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = path[offset:]
		ref, err := r.LookupReference(path)
		if err == nil {
			processed[path] = true
			return callback(ref)
		}
		return nil // ignore error
	})
	if err != nil {
		return err
	}
	refDb := r.NewRefDb()
	refs, err := refDb.GetPackedReferences()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if !processed[ref.name] {
			err = callback(ref)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repository) ForEachGlobReferenceName(pattern string, callback ForEachReferenceNameCallback) error {
	rootDir := filepath.Join(r.pathRepository, GitRefsDir)
	processed := make(map[string]bool)
	offset := len(r.pathRepository)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = path[offset:]
		processed[path] = true
		if fnMatch(pattern, path, 0) {
			return callback(path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	refDb := r.NewRefDb()
	refs, err := refDb.GetPackedReferences()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if !processed[ref.name] {
			if fnMatch(pattern, ref.name, 0) {
				err = callback(ref.name)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Repository) ForEachGlobReference(pattern string, callback ForEachReferenceCallback) error {
	rootDir := filepath.Join(r.pathRepository, GitRefsDir)
	processed := make(map[string]bool)
	offset := len(r.pathRepository)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = path[offset:]
		processed[path] = true
		if fnMatch(pattern, path, 0) {
			ref, err := r.LookupReference(path)
			if err == nil {
				return callback(ref)
			}
			return nil // ignore error
		}
		return nil
	})
	if err != nil {
		return err
	}
	refDb := r.NewRefDb()
	refs, err := refDb.GetPackedReferences()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if !processed[ref.name] {
			if fnMatch(pattern, ref.name, 0) {
				err = callback(ref)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Reference type and its methods
type Reference struct {
	refType        ReferenceType
	repo           *Repository
	targetSymbolic string
	targetOid      *Oid
	name           string
}

func (r *Reference) Target() *Oid {
	return r.targetOid
}

func (r *Reference) SymbolicTarget() string {
	return r.targetSymbolic
}

func (r *Reference) Owner() *Repository {
	return r.repo
}

func (v *Reference) Cmp(ref2 *Reference) int {
	return 0
}

func (r *Reference) Name() string {
	return r.name
}

func (r *Reference) Type() ReferenceType {
	return r.refType
}

func (r *Reference) IsBranch() bool {
	return false
}

func (r *Reference) IsRemote() bool {
	return false
}

func (r *Reference) IsTag() bool {
	return false
}

func (r *Reference) Resolve() (*Reference, error) {
	if r.refType == ReferenceOid {
		return r, nil
	} else {
		return referenceLookupResolved(r.repo, r.targetSymbolic, -1)
	}
}

/*type ReferenceIterator struct {
	repo *Repository
}

func (repo *Repository) NewReferenceIterator() (*ReferenceIterator, error) {

}

func (repo *Repository) NewReferenceIteratorGlob(glob string) (*ReferenceIterator, error) {

}

func (v *ReferenceIterator) Next() (*Reference, error) {

}

type ReferenceNameIterator struct {
	repo *Repository
}

func (repo *Repository) NewReferenceNameIterator() (*ReferenceNameIterator, error) {

}

func (i *ReferenceIterator) Names() *ReferenceNameIterator {
	return &ReferenceNameIterator{i}
}

func (v *ReferenceNameIterator) Next() (string, error) {

}*/

// internal functions

func referenceLookupResolved(repo *Repository, name string, maxNesting int) (*Reference, error) {
	if maxNesting > MaxNestingLevel {
		maxNesting = MaxNestingLevel
	} else if maxNesting < 0 {
		maxNesting = DefaultNestingLevel
	}

	scanType := ReferenceSymbolic
	config := repo.Config()
	precomposeUnicode, _ := config.LookupBool("core.precomposeunicode")
	scanName, err := referenceNormalize(name, precomposeUnicode, true)
	if err != nil {
		return nil, err
	}
	var ref *Reference
	refDb := repo.NewRefDb()

	for nesting := maxNesting; nesting >= 0 && scanType == ReferenceSymbolic; nesting-- {
		if nesting != maxNesting {
			scanName = ref.targetSymbolic
		}
		ref, err = refDb.Lookup(scanName)
		if err != nil {
			return nil, err
		}
		scanType = ref.refType
	}

	if scanType != ReferenceOid && maxNesting != 0 {
		return nil, errors.New(fmt.Sprintf("Cannot resolve reference (>%u levels deep)", maxNesting))
	}
	return ref, nil
}

func referenceNormalize(name string, precomposeUnicode, allowOneLevel bool) (string, error) {
	invalid := false
	if len(name) == 0 {
		invalid = true
	} else if name[0] == '/' {
		invalid = true
	} else {
		name = filepath.Clean(name)
		lastChar := name[len(name)-1]
		if lastChar == '.' || lastChar == '/' {
			invalid = true
		}
		if strings.IndexByte(name, '/') == -1 && !allowOneLevel {
			invalid = true
		}
	}
	if invalid {
		return "", errors.New(fmt.Sprintf("The given reference name '%s' is not valid", name))
	}
	if precomposeUnicode {
		name = norm.NFC.String(name)
	}
	return name, nil
}
