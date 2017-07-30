package git4go

/*
import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"fmt"
)

type RevparseFlag int

const (
	RevparseNone      RevparseFlag = 0
	RevparseSingle    RevparseFlag = 1
	RevparseRange     RevparseFlag = 2
	RevparseMergeBase RevparseFlag = 4
)

type Revspec struct {
	to    Object
	from  Object
	flags RevparseFlag
}

func (rs *Revspec) To() Object {
	return rs.to
}

func (rs *Revspec) From() Object {
	return rs.from
}

func (rs *Revspec) Flags() RevparseFlag {
	return rs.flags
}

func (rs *Revspec) IsSingle() bool {
	return rs.flags == RevparseSingle
}

func (rs *Revspec) IsMergeBase() bool {
	return rs.flags == RevparseMergeBase
}

func objectFromReference(reference *Reference) (Object, error) {
	resolved, err := reference.Resolve()
	if err != nil {
		return nil, err
	}
	return reference.repo.Lookup(resolved.Target())
}

func maybeSha(repo *Repository, spec string) (Object, error) {
	if len(spec) != GitOidHexSize {
		return nil, errors.New("not found")
	}
	oid, err := NewOid(spec)
	if err != nil {
		return nil, err
	}
	return repo.Lookup(oid)
}

func maybeAbbrev(repo *Repository, spec string) (Object, error) {
	oid, err := NewOidFromPrefix(spec)
	if err != nil {
		return nil, err
	}
	return repo.LookupPrefix(oid, len(spec))
}

var describeRegexp *regexp.Regexp

func maybeDescribe(repo *Repository, spec string) (Object, error) {
	if describeRegexp == nil {
		describeRegexp = regexp.MustCompilePOSIX(".+-[0-9]+-g[0-9a-fA-F]+")
	}
	index := strings.Index(spec, "-g")
	if index == -1 {
		return nil, errors.New("not found")
	}
	if !describeRegexp.MatchString(spec) {
		return nil, errors.New("not found")
	}
	return maybeAbbrev(repo, spec[index+2:])
}

func revParseLookupObject(repo *Repository, spec string) (Object, *Reference, error) {
	object, err := maybeSha(repo, spec)
	if err == nil {
		return object, nil, nil
	}
	ref, err := repo.DwimReference(spec)
	if err == nil {
		object, err = repo.Lookup(ref.Target())
		if err == nil {
			return object, ref, nil
		}
	}
	if len(spec) < GitOidHexSize {
		object, err := maybeAbbrev(repo, spec)
		if err == nil {
			return object, nil, nil
		}
	}
	object, err = maybeDescribe(repo, spec)
	return object, nil, err
}

func ensureBaseRevLoaded(object Object, reference *Reference, spec string, identifierLength int, repo *Repository, allowEmptyIdentifier bool) (Object, *Reference, error) {
	if object != nil {
		return object, reference, nil
	}
	if reference != nil {
		object, err := objectFromReference(reference)
		if err != nil {
			return nil, nil, err
		}
		return object, reference, nil
	}
	if !allowEmptyIdentifier && identifierLength == 0 {
		return nil, nil, errors.New("Invalid Spec")
	}
	return revParseLookupObject(repo, spec[:identifierLength])
}

func extractCurlyBracesContent(spec string, pos int) (string, int, error) {
	c := spec[pos]
	if c != '^' && c != '@' {
		return "", 0, errors.New("assertion error: extractCurlyBracesContent()")
	}
	pos++
	if pos >= len(spec) || spec[pos] != '{' {
		return "", 0, errors.New("invalid spec: extractCurlyBracesContent()")
	}
	pos++
	endPos := pos
	specLength := len(spec)
	for endPos < specLength {
		if spec[endPos] == '}' {
			return spec[pos:endPos], endPos + 1, nil
		}
		endPos++
	}
	return "", 0, errors.New("invalid spec: extractCurlyBracesContent()")
}

func deferenceToNonTag(object Object) (Object, error) {
	if object.Type() == ObjectTag {
		return object.Peel(ObjectAny)
	}
	return object, nil
}

func handleCaretCurlySyntax(object Object, curlyBracesContent string) (Object, error) {
	if len(curlyBracesContent) == 0 {
		return deferenceToNonTag(object)
	}
	if curlyBracesContent[0] == '/' {
		return handleGrepSyntax(object.Owner(), object.Id(), curlyBracesContent[1:])
	}
	expectedType := TypeString2Type(curlyBracesContent)
	if expectedType == ObjectBad {
		return nil, errors.New("Invalid Spec: handleCaretCurlySyntax()")
	}
	return object.Peel(expectedType)
}

func handleCaretParentSyntax(baseRev Object, n int) (Object, error) {
	peeled, err := baseRev.Peel(ObjectCommit)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return peeled, nil
	}
	commit := peeled.(*Commit)
	return commit.Parent(n - 1), nil
}

func isDigit(c byte) bool {
	return 0 <= c && c <= 9
}

func strtol(str string, start, base int) (int, int, error) {
	end := start
	for end < len(str) {
		if !isDigit(str[end]) {
			result, err := strconv.ParseInt(str[start:end], base, 64)
			if err != nil {
				return 0, 0, err
			}
			return int(result), end, nil
		}
		end++
	}
	return 0, 0, errors.New("invalid format: strtol()")
}

func extractHowMany(spec string, pos int) (int, int, error) {
	parsed := 0
	accumulated := 0
	kind := spec[pos]
	specLength := len(spec)
	if kind != '^' && kind != '~' {
		return 0, 0, errors.New("assertion error: extractHowMany()")
	}
	for {
		for {
			pos++
			accumulated++
			fmt.Println("@@@", pos, spec, specLength)
			if pos >= specLength-1 || spec[pos] != kind || kind != '~' {
				break
			}
		}
		if pos < specLength && isDigit(spec[pos]) {
			var err error
			parsed, pos, err = strtol(spec, pos, 10)
			if err != nil {
				return 0, 0, err
			}
			accumulated += parsed - 1
		}
		if pos >= specLength-1 || spec[pos] != kind || kind != '~' {
			break
		}
	}
	return accumulated, pos, nil
}

func handlingLinearSyntax(obj Object, n int) (*Commit, error) {
	peeled, err := obj.Peel(ObjectCommit)
	if err != nil {
		return err
	}
	commit := peeled.(*Commit)
	return commit.NthGenAncestor(n)
}

func extractPath(spec string, pos int) (string, int, error) {
	if spec[pos] != ':' {
		return "", 0, errors.New("assertion error: extractPath()")
	}
	return spec[pos+1:], len(spec), nil
}

func handleAtSyntax(spec string, identifierLength int, repo *Repository, curlyBasedContent string) (Object, *Reference, error) {
	parsed, err := strconv.ParseInt(curlyBasedContent, 10, 64)
	isNumeric := (err == nil)
	if curlyBasedContent[0] == '-' && (!isNumeric || parsed == 0) {
		return nil, nil, errors.New("invalid spec")
	}
	if isNumeric {
		if parsed < 0 {
			return retrievePreviouslyCheckedOutBranchOrRevision(repo, spec, -int(parsed))
		} else {
			return retrieveRevObjectFromRefLog(repo, spec, int(parsed))
		}
	}

	if curlyBasedContent == "u" || curlyBasedContent == "upstream" {
		ref, err := retrieveRemoteTrackingReference(repo, spec)
		return nil, ref, err
	}
	// todo: parse time
	// return retrieveRevObjectFromRefLog(repo, spec, timestamp)
	return nil, nil, errors.New("not implemented")
}

func retrievePreviouslyCheckedOutBranchOrRevision(repo *Repository, identifier string, parsed int) (Object, *Reference, error) {

}

func retrieveRevObjectFromRefLog(repo *Repository, identifier string, parsed int) (Object, *Reference, error) {

}

func retrieveRemoteTrackingReference(repo *Repository, identifier string) (*Reference, error) {

}

func ensureLeftHandIdentifierIsNotKnownYet(baseRev Object, reference *Reference) error {
}

func ensureBaseRevIsNotKnownYet(object Object) error {
	if object == nil {
		return nil
	}
	return errors.New("invalid spec: ensureBaseRevIsNotKnownYet()")
}

func anyLeftHandIdentifier(object Object, reference *Reference, identifierLength int) bool {
	if object != nil {
		return true
	}
	if reference != nil {
		return true
	}
	return identifierLength > 0
}

func handleColonSyntax(object Object, path string) (Object, error) {
	peeled, err := object.Peel(ObjectTree)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return peeled, err
	}
	tree := peeled.(*Tree)
	entry, err := tree.EntryByPath(path)
	if err != nil {
		return nil, err
	}
	return tree.repo.Lookup(entry.Id)
}

func handleGrepSyntax(repo *Repository, specOid *Oid, pattern string) (Object, error) {
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	walk, err := repo.Walk()
	if err != nil {
		return nil, err
	}
	walk.Sorting(SortTime)
	if specOid == nil {
		err = walk.PushRef("refs/*")
	} else {
		err = walk.Push(specOid)
	}
	if err != nil {
		return nil, err
	}
	oid := new(Oid)
	for {
		err = walk.Next(oid)
		if err != nil {
			break
		}
		commit, err := repo.LookupCommit(oid)
		if err != nil {
			return nil, err
		}
		if rx.MatchString(commit.Message()) {
			return commit, nil
		}
	}
	if IsErrorCode(err, ErrIterOver) {
		return nil, errors.New("not found")
	}
	return nil, err
}

func (r *Repository) RevparseExt(spec string) (Object, *Reference, error) {
	pos := 0
	identifierLength := 0
	shouldReturnReference := true
	var reference *Reference
	var baseRev Object
	var err error

	for pos < len(spec) {
		switch spec[pos] {
			case '^':
			shouldReturnReference = false
			baseRev, reference, err = ensureBaseRevLoaded(baseRev, reference, spec, identifierLength, r, false)
			if err != nil {
				return nil, nil, err
			}
			if pos+1<len(spec) && spec[pos+1] == '{' {
				var buf string
				buf, pos, err = extractCurlyBracesContent(spec, pos)
				if err != nil {
					return nil, nil, err
				}
				baseRev, err = handleCaretCurlySyntax(baseRev, buf)
			} else {
				var n int
				n, pos, err = extractHowMany(spec, pos)
				if err != nil {
					return nil, nil, err
				}
				baseRev, err = handleCaretParentSyntax(baseRev, n)
				if err != nil {
					return nil, nil, err
				}
			}
			case '~':
			var n int
			shouldReturnReference = false
			n, pos, err = extractHowMany(spec, pos)
			if err != nil {
				return nil, nil, err
			}
			baseRev, reference, err = ensureBaseRevLoaded(baseRev, reference, spec, identifierLength, r, false)
			if err != nil {
				return nil, nil, err
			}
			err = handlingLinearSyntax(baseRev, n)
			if err != nil {
				return nil, nil, err
			}
			case ':':
			shouldReturnReference = false
			var buf string
			buf, pos, err = extractPath(spec, pos)
			if err != nil {
				return nil, nil, err
			}
			if anyLeftHandIdentifier(baseRev, reference, identifierLength) {
				baseRev, reference, err = ensureBaseRevLoaded(baseRev, reference, spec, identifierLength, r, true)
				if err != nil {
					return nil, nil, err
				}
				baseRev, err = handleColonSyntax(baseRev, buf)
				if err != nil {
					return nil, nil, err
				}
			} else {
				if buf[0] == '/' {
					baseRev, err = handleGrepSyntax(r, nil, buf[1:])
					if err != nil {
						return nil, nil, err
					} else {

					}
				} else {
					// todo: support merge-stage path lookup (":2:Makefile")
					// and plain index blob lookup (:i-am/a/blob)
					return nil, nil, errors.New("Unimplemented")
				}

			}
			case '@':
			if pos+1<len(spec) && spec[pos+1] == '{' {
				var buf string
				buf, pos, err = extractCurlyBracesContent(spec, pos)
				if err != nil {
					return nil, nil, err
				}
				err = ensureBaseRevIsNotKnownYet(baseRev)
				if err != nil {
					return nil, nil, err
				}
				baseRev, reference, err = handleAtSyntax(spec, identifierLength, r, buf)
				if err != nil {
					return nil, nil, err
				}
			} else {
				// fall through
			}
			default:
			err = ensureLeftHandIdentifierIsNotKnownYet(baseRev, reference)
			if err != nil {
				return nil, nil, err
			}
			pos++
			identifierLength++
		}
	}
	baseRev, reference, err = ensureBaseRevLoaded(baseRev, reference, spec, identifierLength, r, false)
	if err != nil {
		return nil, nil, err
	}
	if !shouldReturnReference {
		reference = nil
	}
	return baseRev, reference, nil
}

func (r *Repository) RevparseSingle(spec string) (Object, error) {
	obj, _, err := r.RevparseExt(spec)
	return obj, err
}

func (r *Repository) Revparse(spec string) (*Revspec, error) {
	dotdotPos := strings.Index(spec, "..")
	revspec := &Revspec{}

	var err error

	if dotdotPos != -1 {
		leftStr := spec[:dotdotPos]
		var rightStr string
		revspec.flags = RevparseRange
		if spec[dotdotPos+2] == '.' {
			rightStr = spec[dotdotPos+3:]
		} else {
			revspec.flags |= RevparseMergeBase
			rightStr = spec[dotdotPos+2:]
		}
		revspec.from, err = r.RevparseSingle(leftStr)
		if err != nil {
			return nil, err
		}
		revspec.to, err = r.RevparseSingle(rightStr)
		if err != nil {
			return nil, err
		}
	} else {
		revspec.flags = RevparseSingle
		revspec.from, err = r.RevparseSingle(spec)
		if err != nil {
			return nil, err
		}
	}
	return revspec, err
}
*/
