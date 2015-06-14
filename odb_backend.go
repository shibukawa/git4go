package git4go

import (
	"os"
)

type OdbBackendBase struct {
	priority    int
	isAlternate bool
	fileInfo    os.FileInfo
}

func (b *OdbBackendBase) InitBackend(priority int, isAlternate bool, fileInfo os.FileInfo) {
	b.priority = priority
	b.isAlternate = isAlternate
	b.fileInfo = b.fileInfo
}

func (b *OdbBackendBase) Priority() int {
	return b.priority
}

func (b *OdbBackendBase) SameDirectory(info os.FileInfo) bool {
	return os.SameFile(b.fileInfo, info)
}

type OdbBackend interface {
	InitBackend(priority int, isAlternate bool, fileInfo os.FileInfo)
	Priority() int
	SameDirectory(info os.FileInfo) bool
	Read(oid *Oid) (*OdbObject, error)
	ReadPrefix(oid *Oid, length int) (*Oid, *OdbObject, error)
	ReadHeader(oid *Oid) (ObjectType, int64, error)
	Write(objectType ObjectType, oid *Oid, data []byte) error
	Exists(oid *Oid) bool
	ExistsPrefix(oid *Oid, length int) (*Oid, error)
	Refresh()
}

type OdbBackends []OdbBackend

func (a OdbBackends) Len() int           { return len(a) }
func (a OdbBackends) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OdbBackends) Less(i, j int) bool { return a[i].Priority() < a[j].Priority() }
