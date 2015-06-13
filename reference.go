package git4go

import (
	"errors"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"path/filepath"
	"strings"
)

type ReferenceType int

const (
	ReferenceOid          ReferenceType = 1
	ReferenceSymbolic     ReferenceType = 2
	DEFAULT_NESTING_LEVEL               = 5
	MAX_NESTING_LEVEL                   = 10
	GIT_REFNAME_MAX                     = 1024
)

// Repository methods related to Reference

func (r *Repository) LookupReference(name string) (*Reference, error) {
	return referenceLookupResolved(r, name, 0)
}

func (r *Repository) Head() (*Reference, error) {
	head, err := r.LookupReference(GIT_HEAD_FILE)
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
		name = GIT_HEAD_FILE
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

// Reference type and its methods
type Reference struct {
	refType        ReferenceType
	repo           *Repository
	targetSymbolic string
	targetOid      *Oid
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

func (r *Reference) name() string {
	return ""
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

// internal functions

func referenceLookupResolved(repo *Repository, name string, maxNesting int) (*Reference, error) {
	if maxNesting > MAX_NESTING_LEVEL {
		maxNesting = MAX_NESTING_LEVEL
	} else if maxNesting < 0 {
		maxNesting = DEFAULT_NESTING_LEVEL
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
