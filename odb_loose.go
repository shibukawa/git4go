package git4go

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type OdbBackendLoose struct {
	OdbBackendBase
	objectsDir string
	dirMode    uint32
	fileMode   uint32
	doFileSync bool
}

func NewOdbBackendLoose(objectsDir string, compressionLevel int, doFileSync bool, dirMode, fileMode uint32) *OdbBackendLoose {
	if compressionLevel < 0 {
		compressionLevel = zlib.BestSpeed
	}
	if dirMode == 0 {
		dirMode = GIT_OBJECT_DIR_MODE
	}
	if fileMode == 0 {
		fileMode = GIT_OBJECT_FILE_MODE
	}
	return &OdbBackendLoose{
		objectsDir: objectsDir,
		dirMode:    dirMode,
		fileMode:   fileMode,
		doFileSync: doFileSync,
	}
}

func isZlibCompressedData(data []byte) bool {
	w := uint(data[0])<<8 + uint(data[1])
	return (data[0]&0x8F) == 0x08 && (w%31) == 0
}

func parseObjectHeader(data []byte) (ObjectType, int64, int, error) {
	resultType := ObjectBad
	var size int64
	offset := 0
	typeEnd := 0
	var err error
	for ; offset < len(data); offset++ {
		if data[offset] == ' ' {
			resultType = TypeString2Type(string(data[:offset]))
			typeEnd = offset + 1
			offset++
			break
		}
	}
	for ; offset < len(data); offset++ {
		if data[offset] == 0 {
			size, err = strconv.ParseInt(string(data[typeEnd:offset]), 10, 64)
			if err != nil {
				return ObjectBad, -1, 0, err
			}
			offset++
			break
		}
	}
	return resultType, size, offset, nil
}

func parseBinaryObjectHeader(data []byte) (ObjectType, int64, int, error) {
	if len(data) == 0 {
		return ObjectBad, 0, 0, errors.New("parseBinaryObjectHeader: input is empty")
	}
	c := int(data[0])
	resultType := ObjectType((c >> 4) & 7)
	var size int64 = int64(c & 15)
	var shift uint = 4
	offset := 1
	for (c & 0x80) != 0 {
		if len(data) <= offset {
			return ObjectBad, 0, 0, errors.New("parseBinaryObjectHeader: input is too short")
		}
		s := int64(data[offset])
		offset++
		size += (s & 0x7f) << shift
		shift += 7
	}
	return resultType, size, offset, nil
}

func (o *OdbBackendLoose) Read(oid *Oid) (*OdbObject, error) {
	dirName, fileName := oid.PathFormat()
	content, err := ioutil.ReadFile(filepath.Join(o.objectsDir, dirName, fileName))
	if err != nil {
		return nil, err
	}
	if isZlibCompressedData(content) {
		reader, err := zlib.NewReader(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		io.Copy(&buffer, reader)
		data := buffer.Bytes()
		objType, _, offset, err := parseObjectHeader(data)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return &OdbObject{
			Type: objType,
			Data: data[offset:],
		}, nil
	} else {
		objType, _, offset, err := parseBinaryObjectHeader(content)
		if err != nil {
			return nil, err
		}
		reader, err := zlib.NewReader(bytes.NewReader(content[offset:]))
		defer reader.Close()
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		io.Copy(&buffer, reader)
		return &OdbObject{
			Type: objType,
			Data: buffer.Bytes(),
		}, nil
	}
}

func (o *OdbBackendLoose) ReadPrefix(oid *Oid, length int) (*Oid, *OdbObject, error) {
	foundId, err := o.ExistsPrefix(oid, length)
	if err != nil {
		return nil, nil, err
	}
	obj, err := o.Read(foundId)
	if err != nil {
		return nil, nil, err
	}
	return foundId, obj, nil
}

func (o *OdbBackendLoose) ReadHeader(oid *Oid) (ObjectType, int64, error) {
	dirName, fileName := oid.PathFormat()
	content, err := ioutil.ReadFile(filepath.Join(o.objectsDir, dirName, fileName))
	if err != nil {
		return ObjectBad, -1, err
	}
	if isZlibCompressedData(content) {
		reader, err := zlib.NewReader(bytes.NewReader(content))
		if err != nil {
			return ObjectBad, -1, err
		}
		var buffer bytes.Buffer
		io.CopyN(&buffer, reader, 64)
		data := buffer.Bytes()
		objType, size, _, err := parseObjectHeader(data)
		if err != nil {
			return ObjectBad, -1, err
		}
		return objType, size, nil
	} else {
		objType, size, _, err := parseBinaryObjectHeader(content)
		if err != nil {
			return ObjectBad, -1, err
		}
		return objType, size, nil
	}
}

func (o *OdbBackendLoose) Write(data []byte, objType ObjectType) (*Oid, error) {
	oid, err := hash(data, objType)
	if err != nil {
		return nil, err
	}
	dirName, fileName := oid.PathFormat()
	dirPath := filepath.Join(o.objectsDir, dirName)
	os.MkdirAll(dirPath, os.FileMode(GIT_OBJECT_DIR_MODE))
	file, err := os.OpenFile(filepath.Join(dirPath, fileName), os.O_WRONLY, os.FileMode(GIT_OBJECT_FILE_MODE))
	defer file.Close()
	writer := zlib.NewWriter(file)
	fmt.Fprintf(writer, "%s %d\x00", objType.String(), len(data))
	writer.Write(data)
	defer writer.Close()

	return oid, nil
}

func (o *OdbBackendLoose) Exists(oid *Oid) bool {
	dirName, fileName := oid.PathFormat()
	_, err := os.Stat(filepath.Join(o.objectsDir, dirName, fileName))
	return !os.IsNotExist(err)
}

func (o *OdbBackendLoose) ExistsPrefix(oid *Oid, length int) (*Oid, error) {
	dirName, fileName := oid.PathFormat()
	prefix := fileName[:length-2]
	file, err := os.Open(filepath.Join(o.objectsDir, dirName))
	if err != nil {
		return nil, err
	}
	found := 0
	var foundId string
	dirChildNames, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	for _, dirChildName := range dirChildNames {
		if strings.HasPrefix(dirChildName, prefix) {
			found++
			foundId = dirChildName
		}
	}
	if found == 0 {
		return nil, errors.New("no matching loose object for prefix")
	} else if found == 1 {
		return NewOid(dirName + foundId)
	} else {
		return nil, errors.New("multiple matches in loose objects")
	}
}
