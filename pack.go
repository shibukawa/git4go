package git4go

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/edsrzf/mmap-go"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"
)

type PackFile struct {
	lock     sync.RWMutex
	mwf      MWindowFile
	indexMap mmap.MMap

	numObjects   int
	oids         []*Oid
	badObjects   []*Oid
	mtime        time.Time
	packLocal    bool
	packKeep     bool
	hasCache     bool
	indexVersion int

	packName string
	baseName string
}

func (p *PackFile) findEntry(shortOid *Oid, length int) (*PackEntry, bool, error) {
	if length == GIT_OID_HEXSZ {
		for _, badObject := range p.badObjects {
			if shortOid.Equal(badObject) {
				return nil, false, errors.New("bad object found in packfile")
			}
		}
	}
	offset, foundOid, notFound, err := p.findOffset(shortOid, length)
	if err != nil {
		return nil, notFound, err
	}
	if p.mwf.file == nil {
		err = p.open()
		if err != nil {
			return nil, false, err
		}
	}
	return &PackEntry{
		Offset:   offset,
		Sha1:     foundOid,
		PackFile: p,
	}, false, nil
}

func (p *PackFile) findOffset(shortOid *Oid, length int) (uint64, *Oid, bool, error) {
	found := false
	current := 0

	if p.indexVersion == -1 {
		err := p.openIndex()
		if err != nil {
			return 0, nil, false, err
		}
	}
	level1 := *(*[]uint32)(unsafe.Pointer(&p.indexMap))
	level1Offset := 0
	index := 0

	if p.indexVersion > 1 {
		level1Offset = 2
		index = 8
	}
	index += 4 * 256
	firstId := (int)((shortOid)[0])
	hi := ntohl(level1[level1Offset+firstId])
	var lo uint32
	if firstId != 0 {
		lo = ntohl(level1[level1Offset+firstId-1])
	}
	stride := 0
	if p.indexVersion > 1 {
		stride = 20
	} else {
		stride = 24
		index += 4
	}
	var foundOid *Oid
	pos := sha1Position(p.indexMap[index:], stride, lo, hi, shortOid[:])
	if pos >= 0 {
		found = true
		foundOid = new(Oid)
		current = index + pos*stride
		copy(foundOid[:], p.indexMap[current:])
	} else {
		pos = -1 - pos
		if pos < p.numObjects {
			current = index + pos*stride
			foundOid := new(Oid)
			copy(foundOid[:], p.indexMap[current:])
			if shortOid.NCmp(foundOid, uint(length)) == 0 {
				found = true
			}
		}
	}
	if !found {
		return 0, nil, true, errors.New("failed to find offset for pack entry: " + shortOid.String())
	}
	if length != GIT_OID_HEXSZ && pos+1 < p.numObjects {
		next := new(Oid)
		copy(next[:], p.indexMap[current+stride:])
		if shortOid.NCmp(next, uint(length)) == 0 {
			return 0, nil, false, errors.New("found multiple offsets for pack entry")
		}
	}
	offsetOut := p.nthPackedObjectOffset(pos)
	foundOid = NewOidFromBytes(p.indexMap[current:])
	return offsetOut, foundOid, false, nil
}

func sha1Position(table []byte, stride int, lo, hi uint32, key []byte) int {
	for {
		mi := int((lo + hi) / 2)
		cmp := bytes.Compare(table[mi*stride:mi*stride+20], key)
		if cmp == 0 {
			return mi
		}
		if cmp > 0 {
			hi = uint32(mi)
		} else {
			lo = uint32(mi + 1)
		}
		if lo >= hi {
			break
		}
	}
	return -int(lo) - 1
}

func (p *PackFile) nthPackedObjectOffset(n int) uint64 {
	buffer := *(*[]uint32)(unsafe.Pointer(&p.indexMap))

	offset := 256
	if p.indexVersion == 1 {
		return uint64(ntohl(buffer[offset+6*n]))
	} else {
		var off uint32 = 0
		offset += 2 + p.numObjects*6
		off = ntohl(buffer[offset+n])
		if off&0x80000000 == 0 {
			return uint64(off)
		}
		offset += p.numObjects + int((off&0x7fffffff)*2)
		return uint64(ntohl(buffer[offset]))<<32 + uint64(ntohl(buffer[offset+1]))
	}
}

func (p *PackFile) open() error {
	if p.indexVersion == -1 && p.openIndex() != nil {
		return errors.New("failed to open packfile (0)")
	}
	if p.mwf.file != nil {
		return nil
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	file, err := os.Open(p.packName)
	defer file.Close()
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	p.mwf.register()
	if p.mwf.size == 0 {
		if !stat.Mode().IsRegular() {
			return errors.New("failed to open packfile (1)")
		}
		p.mwf.size = stat.Size()
	} else if p.mwf.size != stat.Size() {
		return errors.New("failed to open packfile (2)")
	}
	var hdr_signature uint32
	var hdr_version uint32
	var hdr_entities uint32

	binary.Read(file, binary.BigEndian, &hdr_signature)
	binary.Read(file, binary.BigEndian, &hdr_version)
	binary.Read(file, binary.BigEndian, &hdr_entities)

	if hdr_signature != 0x5041434b /*PACK*/ || !versionOk(hdr_version) || p.numObjects != int(hdr_entities) {
		return errors.New("failed to open packfile (3)")
	}
	var sha1 Oid
	var idxSha1 Oid
	_, err = file.Seek(p.mwf.size-GIT_OID_RAWSZ, os.SEEK_SET)
	if err != nil {
		return errors.New("failed to open packfile (4)")
	}
	file.Read(sha1[:])
	copy(idxSha1[:], p.indexMap[len(p.indexMap)-40:])

	if !sha1.Equal(&idxSha1) {
		return errors.New("failed to open packfile (5)")
	}
	return nil
}

func versionOk(version uint32) bool {
	return version == 2 || version == 3
}

func (p *PackFile) checkIndex(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() < (4*256+20+20) {
		return errors.New("Invalid pack index: " + path)
	}
	p.indexMap, err = mmap.Map(file, 0, mmap.RDONLY)

	if err != nil {
		return err
	}
	buffer := bytes.NewReader(p.indexMap)
	var index_signature uint32
	var index_version uint32
	binary.Read(buffer, binary.BigEndian, &index_signature)
	binary.Read(buffer, binary.BigEndian, &index_version)
	if index_signature != 0xff744f63 /* "\377tOc" */ {
		index_version = 1
	} else if index_version < 2 || 2 < index_version {
		p.indexMap.Unmap()
		return errors.New("unsupported index version")
	}
	var nr uint32
	map32 := *(*[]uint32)(unsafe.Pointer(&p.indexMap))
	index := 0
	if index_version > 1 {
		index += 2
	}
	for i := 0; i < 256; i++ {
		n := ntohl(map32[index+i])
		if n < nr {
			p.indexMap.Unmap()
			return errors.New("index is non-monotonic")
		}
		nr = n
	}
	indexSize := stat.Size()
	if index_version == 1 {
		if indexSize != 4*256+int64(nr)*24+20+20 {
			p.indexMap.Unmap()
			return errors.New("index is corrupted")
		}
	} else if index_version == 2 {
		minSize := int64(8 + 4*256 + nr*(20+4+4) + 20 + 20)
		maxSize := minSize

		if nr != 0 {
			maxSize += int64((nr - 1) * 8)
		}
		if indexSize < minSize || indexSize > maxSize {
			p.indexMap.Unmap()
			return errors.New("wrong index size")
		}
	}
	p.numObjects = int(nr)
	p.indexVersion = int(index_version)
	return nil
}

func (p *PackFile) openIndex() error {
	if p.indexVersion > -1 {
		return nil
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.indexVersion == -1 {
		return p.checkIndex(p.baseName + ".idx")
	}
	return nil
}

func NewPackFile(path string) (*PackFile, error) {
	ext := filepath.Ext(path)
	result := &PackFile{
		baseName:     path[:len(path)-4],
		packLocal:    true,
		indexVersion: -1,
	}
	if ext == ".idx" {
		result.packName = result.baseName + ".pack"
		_, err := os.Stat(result.baseName + ".keep")
		result.packKeep = !os.IsNotExist(err)
	} else {
		result.packName = path
	}

	stat, err := os.Stat(result.packName)
	if os.IsNotExist(err) || !stat.Mode().IsRegular() {
		return nil, errors.New("packfile not found")
	}
	result.mtime = stat.ModTime()
	result.mwf.file = nil
	result.mwf.size = stat.Size()

	return result, nil
}

type PackEntry struct {
	Offset   uint64
	Sha1     *Oid
	PackFile *PackFile
}
