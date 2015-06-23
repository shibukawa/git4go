package git4go

import (
	"bytes"
	"encoding/hex"
	"errors"
	"runtime"
	"strings"
)

const (
	GIT_OID_RAWSZ                    = 20
	GitOidHexSize                    = 40
	GitOidMinimumPrefixLength        = 4
	GIT_OBJECT_DIR_MODE       uint32 = 0777
	GIT_OBJECT_FILE_MODE      uint32 = 0444
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
	if len(s) > GitOidHexSize {
		return nil, errors.New("string is too long for oid")
	}
	o := new(Oid)

	slice, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(slice) != 20 {
		return nil, errors.New("Invalid Oid")
	}
	copy(o[:], slice[:20])
	return o, nil
}

func NewOidFromPrefix(s string) (*Oid, error) {
	if len(s) > GitOidHexSize {
		return nil, errors.New("string is too long for oid")
	}
	slice, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	length := len(s)

	shortOid := new(Oid)
	copy(shortOid[:], slice[:(length+1)/2])
	if (length % 2) == 1 {
		shortOid[length/2] &= 0xF0
	}
	return shortOid, nil
}

func (oid *Oid) String() string {
	return hex.EncodeToString(oid[:])
}

func (oid *Oid) PathFormat() (string, string) {
	idString := oid.String()
	return idString[:2], idString[2:]
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
	result := bytes.Compare(oid[:n/2], oid2[:n/2])
	if result == 0 && n%2 == 1 {
		if (oid[n/2+1]^oid2[n/2+1])&0xf0 != 0 {
			return 1
		}
		return 0
	}
	return result
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

func Is64bit() bool {
	return strings.HasSuffix(runtime.GOARCH, "64")
}

// todo: this code doesn't work on spark
func ntohl(input uint32) uint32 {
	return (input&0xff)<<24 | (input&0xff00)<<8 |
		(input&0xff0000)>>8 | (input&0xff000000)>>24
}
