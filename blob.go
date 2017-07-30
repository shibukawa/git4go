package git4go
import (
	"os"
	"errors"
	"path/filepath"
	"io/ioutil"
)

func (r *Repository) LookupBlob(oid *Oid) (*Blob, error) {
	obj, err := objectLookupPrefix(r, oid, GitOidHexSize, ObjectBlob)
	return obj.(*Blob), err
}

func (r *Repository) LookupPrefixBlob(oid *Oid, length int) (*Blob, error) {
	obj, err := objectLookupPrefix(r, oid, length, ObjectBlob)
	return obj.(*Blob), err
}

func (r *Repository) CreateBlobFromBuffer(data []byte) (*Oid, error) {
	odb, err := r.Odb()
	if err != nil {
		return nil, err
	}
	return odb.Write(data, ObjectBlob)
}

func (r *Repository) CreateBlobFromWorkdir(path string) (*Oid, error) {
	oid, _, err := createBlobCreateFromPaths(r, "", path, 0, true)
	return oid, err
}

type BlobChunkCallback func(maxLen int) ([]byte, error)

// todo
func (r *Repository) CreateBlobFromChunks(hintPath string, callback BlobChunkCallback) (*Oid, error) {
	return nil, nil
}

type Blob struct {
	gitObject
	contents []byte
}

func (b *Blob) Type() ObjectType {
	return ObjectBlob
}

func (b *Blob) Peel(targetType ObjectType) (Object, error) {
	return peel(b, targetType)
}

func (b *Blob) Size() int64 {
	return int64(len(b.contents))
}

func (b *Blob) Contents() []byte {
	return b.contents
}

func newBlob(repo *Repository, oid *Oid, contents []byte) *Blob {
	return &Blob{
		contents: contents,
		gitObject: gitObject{
			repo: repo,
			oid:  oid,
		},
	}
}

func createBlobCreateFromPaths(repo *Repository, contentPath, hintPath string, hintMode Filemode, tryLoadFilters bool) (*Oid, os.FileInfo, error) {
	if hintPath == "" && tryLoadFilters {
		return nil, 0, errors.New("Assertion error")
	}
	if contentPath == "" {
		if repo.IsBare() {
			return nil, nil, MakeGitError("Repository should not be bare", ErrBareRepository)
		}
		contentPath = filepath.Join(repo.Workdir(), hintPath)
	}
	stat, err := os.Lstat(contentPath)
	if err != nil {
		return nil, nil, err
	}
	if stat.IsDir() {
		return nil, nil, MakeGitError("Content path should not be dir", ErrDirectory)
	}
	odb, err := repo.Odb()
	if err != nil {
		return nil, nil, err
	}
	var mode Filemode
	if hintMode != 0 {
		mode = hintMode
	} else {
		mode = Filemode(stat.Mode())
	}
	var oid *Oid
	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		targetPath, err := os.Readlink(contentPath)
		if err != nil {
			return nil, nil, err
		}
		oid, err = repo.CreateBlobFromBuffer([]byte(targetPath))
	} else {
		// todo: filter
		content, err := ioutil.ReadFile(contentPath)
		if err != nil {
			return nil, nil, err
		}
		oid, err = repo.CreateBlobFromBuffer(content)
	}
	if err != nil {
		return nil, nil, err
	}
	return oid, stat, nil
}