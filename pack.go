package git4go

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"github.com/edsrzf/mmap-go"
	"io"
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

func (p *PackFile) findOffset(shortOid *Oid, length int) (offsetOut uint64, foundOid *Oid, notFound bool, err error) {
	notFound = true

	if p.indexVersion == -1 {
		err = p.openIndex()
		if err != nil {
			notFound = false
			return
		}
	}
	level1 := *(*[]uint32)(unsafe.Pointer(&p.indexMap))
	level1Offset := 0
	offset := 0

	if p.indexVersion > 1 {
		level1Offset = 2
		offset = 8
	}
	offset += 4 * 256
	firstId := (int)((shortOid)[0])
	hi := ntohl(level1[level1Offset+firstId])
	var lo uint32
	if firstId != 0 {
		lo = ntohl(level1[level1Offset+firstId-1])
	}
	var stride int
	if p.indexVersion > 1 {
		stride = 20
	} else {
		stride = 24
		offset += 4
	}
	pos := sha1Position(p.indexMap[offset:], stride, lo, hi, shortOid[:])
	var current int
	if pos >= 0 {
		notFound = false
		foundOid = new(Oid)
		current = offset + pos*stride
		copy(foundOid[:], p.indexMap[current:])
	} else {
		pos = -1 - pos
		if pos < p.numObjects {
			current = offset + pos*stride
			foundOid := new(Oid)
			copy(foundOid[:], p.indexMap[current:])
			if shortOid.NCmp(foundOid, uint(length)) == 0 {
				notFound = false
			}
		}
	}
	if notFound {
		err = errors.New("failed to find offset for pack entry: " + shortOid.String())
		return
	}
	if length != GIT_OID_HEXSZ && pos+1 < p.numObjects {
		next := new(Oid)
		copy(next[:], p.indexMap[current+stride:])
		if shortOid.NCmp(next, uint(length)) == 0 {
			err = errors.New("found multiple offsets for pack entry")
			return
		}
	}
	offsetOut = p.nthPackedObjectOffset(pos)
	foundOid = NewOidFromBytes(p.indexMap[current:])
	return
}

func sha1Position(table []byte, stride int, lo, hi uint32, key []byte) int {
	for lo < hi {
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

	var err error
	p.mwf.file, err = os.Open(p.packName)
	if err != nil {
		return err
	}
	stat, err := p.mwf.file.Stat()
	if err != nil {
		return err
	}
	p.mwf.register()
	if p.mwf.size == 0 {
		if !stat.Mode().IsRegular() {
			return errors.New("failed to open packfile (1)")
		}
		p.mwf.size = uint64(stat.Size())
	} else if p.mwf.size != uint64(stat.Size()) {
		return errors.New("failed to open packfile (2)")
	}
	var hdr_signature uint32
	var hdr_version uint32
	var hdr_entities uint32

	binary.Read(p.mwf.file, binary.BigEndian, &hdr_signature)
	binary.Read(p.mwf.file, binary.BigEndian, &hdr_version)
	binary.Read(p.mwf.file, binary.BigEndian, &hdr_entities)

	if hdr_signature != 0x5041434b /*PACK*/ || !versionOk(hdr_version) || p.numObjects != int(hdr_entities) {
		return errors.New("failed to open packfile (3)")
	}
	var sha1 Oid
	var idxSha1 Oid
	_, err = p.mwf.file.Seek(int64(p.mwf.size-GIT_OID_RAWSZ), os.SEEK_SET)
	if err != nil {
		return errors.New("failed to open packfile (4)")
	}
	p.mwf.file.Read(sha1[:])
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

func (p *PackFile) openWindow(offset uint64) ([]byte, error) {
	if p.mwf.file == nil {
		err := p.open()
		if err != nil {
			return nil, err
		}
	}
	if offset > (p.mwf.size - 20) {
		return nil, errors.New("invalid size")
	}
	return p.mwf.Open(offset, 20)
}

func (p *PackFile) resolveHeader(offset uint64) (ObjectType, uint64, error) {
	elem, err := p.unpackHeader(offset)
	if err != nil {
		return ObjectBad, 0, err
	}
	var resultSize uint64
	var baseOffset uint64
	objType := elem.objType
	if objType == ObjectOfsDelta || objType == ObjectRefDelta {
		var curPos uint64
		baseOffset, curPos, _ = p.getDeltaBase(elem.offset, elem.objType, offset)
		delta, err := p.unpackCompressed(curPos, elem.objType)
		if err != nil {
			return ObjectBad, 0, err
		}
		_, targetSize, offset := decodeHeader(delta)
		resultSize = targetSize
		curPos += offset
	} else {
		resultSize = elem.size
	}

	for objType == ObjectOfsDelta || objType == ObjectRefDelta {
		elem, err = p.unpackHeader(baseOffset)
		if err != nil {
			return ObjectBad, 0, err
		}
		objType = elem.objType
		if objType != ObjectOfsDelta && objType != ObjectRefDelta {
			break
		}
		baseOffset, _, _ = p.getDeltaBase(elem.offset, objType, baseOffset)
	}
	return elem.objType, resultSize, nil
}

func (p *PackFile) unpackCompressed(offset uint64, objType ObjectType) ([]byte, error) {
	data, err := p.openWindow(offset)
	if err != nil {
		return nil, err
	}
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	io.Copy(&buffer, reader)
	return buffer.Bytes(), nil
}

func (p *PackFile) unpackHeader(curPos uint64) (*PackChainElem, error) {
	buffer, err := p.mwf.Open(curPos, 20)
	if err != nil {
		return nil, err
	}

	c := buffer[0]
	objType := ObjectType((c >> 4) & 7)
	size := uint64(c & 15)
	shift := uint(4)
	length := len(buffer)
	used := 1
	for c&0x80 != 0 {
		if length < used {
			return nil, errors.New("buffer too small")
		}
		if 64 <= shift {
			return nil, errors.New("packfile corrupted")
		}
		c = buffer[used]
		used++
		size += uint64(c&0x7f) << shift
		shift += 7
	}

	result := &PackChainElem{
		objType: objType,
		offset:  curPos + uint64(used),
		size:    size,
	}

	return result, nil
}

func MSB(x uint64, bit uint) bool {
	return (x & ((0xffffffffffffffff) << (64 - bit))) != 0
}

func (p *PackFile) getDeltaBase(curPos uint64, objType ObjectType, deltaObjOffset uint64) (baseOffset, resultCurPos uint64, err error) {
	var buffer []byte
	buffer, err = p.openWindow(curPos)
	if err != nil {
		return 0, 0, err
	}
	resultCurPos = curPos
	if objType == ObjectOfsDelta {
		c := buffer[0]
		baseOffset = uint64(c & 127)
		used := 1
		for c&128 != 0 {
			if len(buffer) <= used {
				err = errors.New("getDeltaBase: buffer over flow(1)")
				return
			}
			baseOffset += 1
			if baseOffset == 0 || MSB(baseOffset, 7) {
				err = errors.New("getDeltaBase: buffer over flow(2)")
				return
			}
			c = buffer[used]
			used++
			baseOffset = (baseOffset << 7) + uint64(c&127)
		}
		baseOffset = deltaObjOffset - baseOffset
		if baseOffset <= 0 || baseOffset >= deltaObjOffset {
			err = errors.New("getDeltaBase: buffer out of bound")
			baseOffset = 0 // out of bound
			return
		}
		resultCurPos += uint64(used)
		return
	} else if objType == ObjectRefDelta {
		if p.hasCache {
			// todo
		}
		baseOffset, _, _, err = p.findOffset(NewOidFromBytes(buffer), GIT_OID_HEXSZ)
		if err != nil {
			return 0, 0, errors.New("base entry delta is not in the same pack")
		}
		resultCurPos += 20
		return
	} else {
		baseOffset = 0
		return
	}
	return
}

func (p *PackFile) dependencyChain(objOffset uint64) (stack []*PackChainElem, resultObjOffset uint64, err error) {
	var baseOffset uint64
	stack = make([]*PackChainElem, 0, 64)
	for {
		var elem *PackChainElem
		elem, err = p.unpackHeader(objOffset)
		if err != nil {
			return
		}
		elem.baseKey = objOffset
		stack = append(stack, elem)
		if elem.objType != ObjectOfsDelta && elem.objType != ObjectRefDelta {
			break
		}

		baseOffset, elem.offset, err = p.getDeltaBase(elem.offset, elem.objType, objOffset)
		if err != nil && baseOffset == 0 {
			err = errors.New("delta offset is zero")
		}
		if err != nil {
			return
		}
		objOffset = uint64(baseOffset)
	}
	return
}

func (p *PackFile) unpack(objOffset uint64) (obj *OdbObject, resultObjOffset uint64, err error) {
	stack, resultObjOffset, err := p.dependencyChain(objOffset)
	if err != nil {
		return
	}
	lastElem := stack[len(stack)-1]
	baseType := lastElem.objType
	obj = &OdbObject{
		Type: baseType,
	}
	var baseData []byte
	if baseType == ObjectCommit || baseType == ObjectTree || baseType == ObjectTag || baseType == ObjectBlob {
		baseData, err = p.unpackCompressed(lastElem.offset, lastElem.objType)
		obj.Data = baseData
		if err != nil {
		}
	} else if baseType == ObjectOfsDelta || baseType == ObjectRefDelta {
		err = errors.New("dependency chain ends in a delta")
		return
	} else {
		err = errors.New("invalid packfile type in header")
		return
	}
	//var buffer bytes.Buffer
	for i := len(stack) - 2; i >= 0; i-- {
		elem := stack[i]
		delta, err := p.unpackCompressed(elem.offset, elem.objType)
		if err != nil {
			err = errors.New("can't read unpack delta")
			continue
		}
		baseData, err = ApplyDelta(baseData, delta)
		if err != nil {
			err = errors.New("can't apply delta")
			continue
		}
		obj.Data = baseData
	}
	return
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
	result.mwf.size = uint64(stat.Size())

	return result, nil
}

func GetPack(path string) (*PackFile, error) {
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

func PutPack(pack *PackFile) error {
	mwindowMutex.Lock()
	defer mwindowMutex.Unlock()
	delete(packCache, pack.packName)
	return nil
}

type PackEntry struct {
	Offset   uint64
	Sha1     *Oid
	PackFile *PackFile
}

type PackChainElem struct {
	baseKey uint64
	objType ObjectType
	offset  uint64
	size    uint64
}
