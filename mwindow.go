package git4go

import (
	"os"
	"sync"
)

type MWindow struct {
}

type MWindowFile struct {
	mwindow *MWindow
	file    *os.File
	size    int64
}

type MWindowCtl struct {
	mapped          int
	openWindow      uint
	mmapCalls       uint
	peakOpenWindows uint
	peakMapped      int
	usedCtr         int
	windowFiles     []*MWindowFile
}

var memCtl MWindowCtl = MWindowCtl{}

var packCache map[string]*PackFile = map[string]*PackFile{}
var mwindowMutex sync.Mutex
var is64bit bool = Is64bit()

func mwindowGetPack(path string) (*PackFile, error) {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()

	existingEntry, ok := packCache[path]
	if ok {
		return existingEntry, nil
	}
	packFile, err := NewPackFile(path)
	if err != nil {
		return nil, err
	}
	packCache[path] = packFile
	return packFile, nil
}

func (mwf *MWindowFile) register() {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()
	memCtl.windowFiles = append(memCtl.windowFiles)
}
