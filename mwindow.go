package git4go

import (
	"errors"
	"github.com/edsrzf/mmap-go"
	"os"
	"runtime"
	"strings"
	"sync"
)

var mwindow_windowSize = 1024 * 1024 * SelectByArch(32, 1024)
var mwindow_mappedLimit = 1024 * 1024 * SelectByArch(256, 8192)
var memCtl MWindowCtl = MWindowCtl{}
var packCache map[string]*PackFile = map[string]*PackFile{}
var mwindowMutex sync.Mutex

type MWindow struct {
	windowMap mmap.MMap
	offset    uint64
	lastUsed  uint64
}

func mwindowFinalizer(w *MWindow) {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()

	w.windowMap.Unmap()
}

func (w *MWindow) contains(offset uint64) bool {
	return w.offset <= offset && offset <= (w.offset+uint64(len(w.windowMap)))
}

type MWindowFile struct {
	windows []*MWindow
	file    *os.File
	size    uint64
}

func (mwf *MWindowFile) Open(offset, extra uint64) ([]byte, error) {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()
	var w *MWindow

	for _, existingWindow := range mwf.windows {
		if existingWindow.contains(offset) && existingWindow.contains(offset+extra) {
			w = existingWindow
			break
		}
	}
	if w == nil {
		var err error
		w, err = mwf.newWindow(offset)
		if err != nil {
			return nil, err
		}
		mwf.windows = append([]*MWindow{w}, mwf.windows...)
	}
	w.lastUsed = memCtl.usedCtr
	memCtl.usedCtr++
	offset -= w.offset
	return w.windowMap[offset:], nil
}

func (mwf *MWindowFile) freeAll() {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()
	mwf.freeAllLocked()
}

func (mwf *MWindowFile) newWindow(offset uint64) (*MWindow, error) {
	wAlign := mwindow_windowSize / 2
	w := &MWindow{
		offset: (offset / wAlign) * wAlign,
	}
	length := mwf.size - w.offset
	if length > mwindow_windowSize {
		length = mwindow_windowSize
	}
	memCtl.mapped += length

	for mwindow_mappedLimit < memCtl.mapped && mwf.closeLru() == nil {
		/* nop */
	}

	mmapObj, err := mmap.MapRegion(mwf.file, int(mwf.size), 0, mmap.RDONLY, int64(w.offset))
	if err != nil {
		return nil, err
	}
	w.windowMap = mmapObj
	runtime.SetFinalizer(w, mwindowFinalizer)
	memCtl.mmapCalls++
	memCtl.openWindow++
	if memCtl.mapped > memCtl.peakMapped {
		memCtl.peakMapped = memCtl.mapped
	}
	if memCtl.openWindow > memCtl.peakOpenWindows {
		memCtl.peakOpenWindows = memCtl.openWindow
	}
	return w, nil
}

func (mwf *MWindowFile) freeAllLocked() {
	for i, w := range memCtl.windowFiles {
		if w == mwf {
			newFiles := memCtl.windowFiles[:i]
			if i != len(memCtl.windowFiles)-1 {
				newFiles = append(newFiles, memCtl.windowFiles[i+1:]...)
			}
			memCtl.windowFiles = newFiles
			break
		}
	}
	for _, window := range mwf.windows {
		memCtl.mapped -= uint64(len(window.windowMap))
		memCtl.openWindow--
		window.windowMap.Unmap()
	}
}

func (mwf *MWindowFile) register() {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()

	memCtl.windowFiles = append(memCtl.windowFiles)
}

func (mwf *MWindowFile) unregister() {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()

	for i, w := range memCtl.windowFiles {
		if w == mwf {
			newFiles := memCtl.windowFiles[:i]
			if i != len(memCtl.windowFiles)-1 {
				newFiles = append(newFiles, memCtl.windowFiles[i+1:]...)
			}
			break
		}
	}
}

func (mwf *MWindowFile) scanLru(lruWindow **MWindow, lruFile **MWindowFile, lruIndex *int) {
	for i, window := range mwf.windows {
		if (*lruWindow) == nil || window.lastUsed < (*lruWindow).lastUsed {
			*lruWindow = window
			*lruFile = mwf
			*lruIndex = i
		}
	}
}

func (mwf *MWindowFile) closeLru() error {
	var lruWindow *MWindow
	var lruFile *MWindowFile
	var lruIndex int

	if len(mwf.windows) != 0 {
		mwf.scanLru(&lruWindow, &lruFile, &lruIndex)
	}
	var currentWindowFile *MWindowFile
	for _, currentWindowFile = range memCtl.windowFiles {
		currentWindowFile.scanLru(&lruWindow, &lruFile, &lruIndex)
	}
	if lruWindow == nil {
		return errors.New("Failed to close memory window. Couldn't find LRU")
	}
	memCtl.mapped -= uint64(len(lruWindow.windowMap))
	lruFile.windows = append(lruFile.windows[:lruIndex], lruFile.windows[lruIndex+1:]...)
	return nil
}

type MWindowCtl struct {
	mapped          uint64
	openWindow      uint64
	mmapCalls       uint64
	peakOpenWindows uint64
	peakMapped      uint64
	usedCtr         uint64
	windowFiles     []*MWindowFile
}

func SelectByArch(for32, for64 uint64) uint64 {
	if strings.HasSuffix(runtime.GOARCH, "64") {
		return for64
	} else {
		return for32
	}
}
