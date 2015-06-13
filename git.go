package git4go

import (
	"bytes"
	"encoding/hex"
	"errors"
)

const (
	GIT_OID_RAWSZ        = 20
	GIT_OID_HEXSZ        = 40
	GIT_OID_MINPREFIXLEN = 4
)

func Discover(start string, acrossFs bool, ceilingDirs []string) (string, error) {
	var flags uint32 = 0
	if acrossFs {
		flags = GIT_REPOSITORY_OPEN_CROSS_FS
	}
	repoPath, _, _, err := findRepo(start, flags, ceilingDirs)
	return repoPath, err
}

type Oid [20]byte

func NewOidFromBytes(b []byte) *Oid {
	oid := new(Oid)
	copy(oid[0:20], b[0:20])
	return oid
}

func NewOid(s string) (*Oid, error) {
	if len(s) > GIT_OID_HEXSZ {
		return nil, errors.New("string is too long for oid")
	}
	o := new(Oid)

	slice, error := hex.DecodeString(s)
	if error != nil {
		return nil, error
	}

	if len(slice) != 20 {
		return nil, errors.New("Invalid Oid")
	}
	copy(o[:], slice[:20])
	return o, nil
}

func (oid *Oid) String() string {
	return hex.EncodeToString(oid[:])
}

func (oid *Oid) Cmp(oid2 *Oid) int {
	return bytes.Compare(oid[:], oid2[:])
}

func (oid *Oid) Copy() *Oid {
	ret := new(Oid)
	copy(ret[:], oid[:])
	return ret
}

func (oid *Oid) Equal(oid2 *Oid) bool {
	return bytes.Equal(oid[:], oid2[:])
}

func (oid *Oid) IsZero() bool {
	for _, a := range oid {
		if a != 0 {
			return false
		}
	}
	return true
}

func (oid *Oid) NCmp(oid2 *Oid, n uint) int {
	return bytes.Compare(oid[:n], oid2[:n])
}

/*func ShortenOids(ids []*Oid, minlen int) (int, error) {
	shorten := C.git_oid_shorten_new(C.size_t(minlen))
	if shorten == nil {
		panic("Out of memory")
	}
	defer C.git_oid_shorten_free(shorten)

	var ret C.int

	for _, id := range ids {
		buf := make([]byte, 41)
		C.git_oid_fmt((*C.char)(unsafe.Pointer(&buf[0])), id.toC())
		buf[40] = 0
		ret = C.git_oid_shorten_add(shorten, (*C.char)(unsafe.Pointer(&buf[0])))
		if ret < 0 {
			return int(ret), MakeGitError(ret)
		}
	}
	return int(ret), nil
}
*/
