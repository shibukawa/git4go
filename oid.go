package git4go

import (
	"bytes"
	"encoding/hex"
	"errors"
)

type Oid [GitOidRawSize]byte

func NewOidFromBytes(b []byte) *Oid {
	oid := new(Oid)
	copy(oid[0:GitOidRawSize], b[0:GitOidRawSize])
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

	if len(slice) != GitOidRawSize {
		return nil, errors.New("Invalid Oid")
	}
	copy(o[:], slice[:GitOidRawSize])
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

func parseOidWithPrefix(buffer []byte, index int, prefix []byte) (*Oid, int) {
	prefixLength := len(prefix)
	if len(buffer)-index < prefixLength+GitOidHexSize+1 {
		return nil, index
	}
	if bytes.Compare(buffer[index:index+prefixLength], prefix) != 0 {
		return nil, index
	}
	if buffer[index+prefixLength+GitOidHexSize] != '\n' {
		return nil, index
	}
	oid, err := NewOid(string(buffer[index+prefixLength : index+prefixLength+GitOidHexSize]))
	if err != nil {
		return nil, index
	}
	return oid, index + prefixLength + GitOidHexSize + 1
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
