package git4go

import (
	"compress/zlib"
	"os"
	"path/filepath"
	"strings"
	"errors"
)

type OdbBackendLoose struct {
	OdbBackendBase
	objectsDir string
	dirMode uint
	fileMode uint
	doFileSync bool

}

func NewOdbBackendLoose(objectsDir string, compressionLevel int, doFileSync bool, dirMode, fileMode uint) *OdbBackendLoose {
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
		dirMode: dirMode,
		fileMode: fileMode,
		doFileSync: doFileSync,
	}
}

func (o *OdbBackendLoose) Read(objectType ObjectType, oid *Oid) ([]byte, int, error) {
	return []byte{}, 0, nil
}

func (o *OdbBackendLoose) ReadPrefix(objectType ObjectType, oid *Oid, length int) ([]byte, int, error) {
	return []byte{}, 0, nil
}

func (o *OdbBackendLoose) ReadHeader(oid *Oid) (int, ObjectType) {
	return 0, ObjectAny
}
func (o *OdbBackendLoose) Write(objectType ObjectType, oid *Oid, data []byte) error {
	return nil
}

func (o *OdbBackendLoose) Exists(oid *Oid) bool {
	dirName, fileName := oid.PathFormat()
	_, err := os.Stat(filepath.Join(o.objectsDir, dirName, fileName))
	return !os.IsNotExist(err)
}

func (o *OdbBackendLoose) ExistsPrefix(oid *Oid, length int) (*Oid, error) {
	dirName, fileName := oid.PathFormat()
	prefix := fileName[:length - 2]
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

func (o *OdbBackendLoose) Refresh() {
}

