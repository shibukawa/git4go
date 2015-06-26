package git4go

import (
	"runtime"
	"strings"
)

const (
	GitOidRawSize                    = 20
	GitOidHexSize                    = 40
	GitOidMinimumPrefixLength        = 4
	GitObjectDirMode          uint32 = 0777
	GitObjectFileMode         uint32 = 0444
)

func Discover(start string, acrossFs bool, ceilingDirs []string) (string, error) {
	var flags uint32 = 0
	if acrossFs {
		flags = GIT_REPOSITORY_OPEN_CROSS_FS
	}
	repoPath, _, _, err := findRepo(start, flags, ceilingDirs)
	return repoPath, err
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
