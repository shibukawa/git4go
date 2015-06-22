package git4go

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

const (
	DELTA_SIZE_MIN = 4
)

func CreateDelta(source, target []byte, maxDeltaSize uint64) ([]byte, error) {
	opcodes := new(bytes.Buffer)
	count := int(math.Ceil(float64(len(source)) / 17.0))
	blocks := NewBlocks(count)

	encodeHeader(opcodes, source, target)

	for i := 0; i < len(source); {
		block := sliceBlock(source, i)
		blocks.set(block, i)
		i += len(block)
	}

	i := 0
	bufferedLength := 0
	insertBuffer := make([]byte, 127)
	for i < len(target) {
		block := sliceBlock(target, i)
		var match *Match = nil
		matchOffsets := blocks.get(block)
		if matchOffsets != nil {
			match = chooseMatch(source, matchOffsets, target, i)
		}
		if match == nil || match.length < DELTA_SIZE_MIN {
			insertLength := len(block)
			if match != nil {
				insertLength += match.length
			}
			if bufferedLength+insertLength <= len(insertBuffer) {
				copy(insertBuffer[bufferedLength:], target[i:i+insertLength])
				bufferedLength += insertLength
			} else {
				err := emitInsert(opcodes, insertBuffer, bufferedLength)
				if err != nil {
					return nil, err
				}
				copy(insertBuffer[:], target[i:i+insertLength])
				bufferedLength = insertLength
			}
			i += insertLength
		} else {
			if bufferedLength > 0 {
				err := emitInsert(opcodes, insertBuffer, bufferedLength)
				if err != nil {
					return nil, err
				}
				bufferedLength = 0
			}
			emitCopy(opcodes, source, match.offset, match.length)
			i += match.length
		}
	}

	if bufferedLength > 0 {
		err := emitInsert(opcodes, insertBuffer, bufferedLength)
		if err != nil {
			return nil, err
		}
		bufferedLength = 0
	}

	if i != len(target) {
		return nil, errors.New("error computing delta buffer")
	} else {
		return opcodes.Bytes(), nil
	}
}

func ApplyDelta(base, delta []byte) ([]byte, error) {
	baseSize, targetSize, offset := decodeHeader(delta)
	if baseSize != uint64(len(base)) {
		return nil, errors.New(fmt.Sprintf("invalid base buffer length in header: %d, %d\n", baseSize, len(base)))
	}
	rv := make([]byte, targetSize)
	var rvOffset uint64

	for offset < uint64(len(delta)) {
		opcode := delta[offset]
		offset++
		if (opcode & 0x80) != 0 {
			var baseOffset uint64
			var copyLength uint64
			if (opcode & 0x01) != 0 {
				baseOffset = uint64(delta[offset])
				offset++
			}
			if (opcode & 0x02) != 0 {
				baseOffset |= uint64(delta[offset]) << 8
				offset++
			}
			if (opcode & 0x04) != 0 {
				baseOffset |= uint64(delta[offset]) << 16
				offset++
			}
			if (opcode & 0x08) != 0 {
				baseOffset |= uint64(delta[offset]) << 24
				offset++
			}
			if (opcode & 0x10) != 0 {
				copyLength = uint64(delta[offset])
				offset++
			}
			if (opcode & 0x20) != 0 {
				copyLength |= uint64(delta[offset]) << 8
				offset++
			}
			if (opcode & 0x40) != 0 {
				copyLength |= uint64(delta[offset]) << 16
				offset++
			}
			if copyLength == 0 {
				copyLength = 0x10000
			}
			copy(rv[rvOffset:], base[baseOffset:baseOffset+copyLength])
			rvOffset += copyLength
		} else if opcode != 0 {
			copyLength := uint64(opcode)
			copy(rv[rvOffset:], delta[offset:offset+copyLength])
			offset += copyLength
			rvOffset += copyLength
		} else {
			return nil, errors.New(fmt.Sprintf("invalid delta opcode at %d\n", offset))
		}
	}
	if rvOffset != targetSize {
		return nil, errors.New("error patching the base buffer")
	}
	return rv, nil
}

// internal functions

type DeltaIndexEntry struct {
	Ptr  int
	Val  uint
	Next *DeltaIndexEntry
}

type DeltaIndex struct {
	MemSize  uint64
	Source   []byte
	HashMask uint32
	Hash     []*DeltaIndexEntry
}

func encodeSize(opcodes *bytes.Buffer, size uint32) {
	lastCode := byte(size & 0x7f)
	size = size >> 7
	for size > 0 {
		opcodes.WriteByte(lastCode | 0x80)
		lastCode = byte(size & 0x7f)
		size = size >> 7
	}
	opcodes.WriteByte(lastCode)
}

func encodeHeader(opcodes *bytes.Buffer, src, target []byte) {
	encodeSize(opcodes, uint32(len(src)))
	encodeSize(opcodes, uint32(len(target)))
}

func sliceBlock(buffer []byte, pos int) []byte {
	j := pos
	for j < len(buffer) && buffer[j] != 10 && (j-pos < 90) {
		j++
	}
	if j < len(buffer) && buffer[j] == 10 {
		j++
	}
	return buffer[pos:j]
}

type Match struct {
	length int
	offset int
}

func chooseMatch(source []byte, sourcePositions []int, target []byte, targetPos int) *Match {
	var rv *Match
	limit := len(source) / 5
	for i, spos := range sourcePositions {
		length := 0
		tpos := targetPos
		if rv != nil && spos < (rv.offset+rv.length) {
			continue
		}
		for tpos < len(target) && spos < len(source) && source[spos] == target[tpos] {
			spos++
			tpos++
			length++
		}
		if rv == nil {
			rv = &Match{
				length: length,
				offset: sourcePositions[i],
			}
		} else if rv.length < length {
			rv.length = length
			rv.offset = sourcePositions[i]
		}
		if rv.length > limit {
			break
		}
	}
	return rv
}

func emitInsert(opcodes *bytes.Buffer, buffer []byte, length int) error {
	if length > 127 {
		return errors.New("invalid insert opcode")
	}
	opcodes.WriteByte(byte(length))
	opcodes.Write(buffer[:length])
	return nil
}

func emitCopy(opcodes *bytes.Buffer, source []byte, offset, length int) {
	ops := make([]byte, 0, 7)
	code := byte(0x80)

	if offset&0xff != 0 {
		ops = append(ops, byte(offset&0xff))
		code |= 0x01
	}
	if offset&0xff00 != 0 {
		ops = append(ops, byte((offset&0xff00)>>8))
		code |= 0x02
	}
	if offset&0xff0000 != 0 {
		ops = append(ops, byte((offset&0xff0000)>>16))
		code |= 0x04
	}
	if offset&0xff000000 != 0 {
		ops = append(ops, byte((offset&0xff000000)>>24))
		code |= 0x08
	}

	if length&0xff != 0 {
		ops = append(ops, byte(length&0xff))
		code |= 0x10
	}
	if length&0xff00 != 0 {
		ops = append(ops, byte((length&0xff00)>>8))
		code |= 0x20
	}
	if length&0xff0000 != 0 {
		ops = append(ops, byte((length&0xff0000)>>16))
		code |= 0x40
	}
	opcodes.WriteByte(code)
	opcodes.Write(ops)
}

func nextSize(buffer []byte, offset uint64) (uint64, uint64) {
	b := buffer[offset]
	offset++
	rv := uint64(b & 0x7f)
	var shift uint = 7

	for (b & 0x80) != 0 {
		b = buffer[offset]
		rv |= uint64(b&0x7f) << shift
		offset++
		shift += 7
	}
	return rv, offset
}

func decodeHeader(buffer []byte) (sourceLength, targetLength, offset uint64) {
	sourceLength, offset = nextSize(buffer, offset)
	targetLength, offset = nextSize(buffer, offset)
	return
}

// internal data structure

func hashBlock(buffer []byte) uint32 {
	var rv uint32 = 0
	j := len(buffer)
	var w uint32 = 1
	for i := 0; i < j; i++ {
		w *= 29
		w = w & 1073741824
		rv += uint32(buffer[i]) * w
		rv = rv & 1073741824
	}
	return rv
}

type Blocks struct {
	n       int
	buckets []Bucket
}

func NewBlocks(n int) *Blocks {
	return &Blocks{
		n:       n,
		buckets: make([]Bucket, n),
	}
}

func (b *Blocks) set(key []byte, value int) {
	index := int(hashBlock(key)) / b.n
	b.buckets[index].set(key, value)

}

func (b *Blocks) get(key []byte) []int {
	index := int(hashBlock(key)) / b.n
	return b.buckets[index].get(key)
}

type Bucket struct {
	keys   [][]byte
	values [][]int
}

func (b *Bucket) set(key []byte, value int) {
	found := false
	for i, storedKey := range b.keys {
		if bytes.Compare(key, storedKey) == 0 {
			b.values[i] = append(b.values[i], value)
			found = true
			break
		}
	}
	if !found {
		b.keys = append(b.keys, key)
		b.values = append(b.values, []int{value})
	}
}

func (b *Bucket) get(key []byte) []int {
	for i, storedKey := range b.keys {
		if bytes.Compare(key, storedKey) == 0 {
			return b.values[i]
		}
	}
	return nil
}
