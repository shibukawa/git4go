package git4go

import (
	"crypto/sha1"
	"runtime"
	"strconv"
	"strings"
)

func Is64bit() bool {
	return strings.HasSuffix(runtime.GOARCH, "64")
}

// todo: this code doesn't work on spark or any big endien arches
func ntohl(input uint32) uint32 {
	return (input&0xff)<<24 | (input&0xff00)<<8 |
		(input&0xff0000)>>8 | (input&0xff000000)>>24
}

func ntohlFromBytes(buffer []byte, offset int) uint32 {
	return uint32(buffer[offset])<<24 | uint32(buffer[offset+1])<<16 |
		uint32(buffer[offset+2])<<8 | uint32(buffer[offset+3])
}

func ntohs(input uint16) uint16 {
	return (input&0xff)<<8 | (input&0xff00)>>8
}

func ntohsFromBytes(buffer []byte, offset int) uint16 {
	return uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
}

func findChar(input []byte, char byte, offset, bound int) int {
	for offset < bound {
		if input[offset] == char {
			return offset
		}
		offset++
	}
	return -1
}

func readString(input []byte, offset, bound int) (string, int) {
	initialOffset := offset
	for offset < bound {
		if input[offset] == 0 {
			str := string(input[initialOffset:offset])
			return str, offset + 1
		}
		offset++
	}
	return "", -1
}

func strtol32(buffer []byte, startOffset, bound, base int) (int64, int) {
	for offset := startOffset; offset < bound; offset++ {
		c := buffer[offset]
		if c < '0' || '9' < c {
			attr, _ := strconv.ParseInt(string(buffer[startOffset:offset]), base, 64)
			return attr, offset + 1
		}
	}
	return -1, startOffset
}

func calcHash(buffer []byte) *Oid {
	h := sha1.New()
	h.Write(buffer)
	sha1Hash := h.Sum(nil)
	oid := new(Oid)
	copy(oid[:], sha1Hash[:])
	return oid
}
