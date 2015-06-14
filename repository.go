package git4go

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	GIT_REPOSITORY_OPEN_NO_FLAG   uint32 = 0
	GIT_REPOSITORY_OPEN_NO_SEARCH uint32 = (1 << 0)
	GIT_REPOSITORY_OPEN_CROSS_FS  uint32 = (1 << 1)
	GIT_REPOSITORY_OPEN_BARE      uint32 = (1 << 2)
	GIT_OBJECTS_DIR               string = "objects/"
	GIT_HEAD_FILE                 string = "HEAD"
	GIT_REFS_DIR                  string = "refs/"
)

// Repository type and its methods

type Repository struct {
	pathRepository string
	workDir        string
	namespace      string
	pathGitLink    string
	isBare         bool
	config         *Config
	refDb          *RefDb
	odb            *Odb
	//cache          *Cache
}

func OpenRepository(path string) (*Repository, error) {
	return openRepository(path, GIT_REPOSITORY_OPEN_NO_SEARCH)
}

func OpenRepositoryExtended(path string) (*Repository, error) {
	return openRepository(path, GIT_REPOSITORY_OPEN_NO_FLAG)
}

func (r *Repository) Path() string {
	return r.pathRepository
}

func (r *Repository) WorkDir() string {
	if r.isBare {
		return ""
	}
	return r.workDir
}

// internal functions

func openRepository(path string, flags uint32) (*Repository, error) {
	path, parent, link_path, err := findRepo(path, flags, []string{})
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		pathRepository: path,
		pathGitLink:    link_path,
		isBare:         (flags & GIT_REPOSITORY_OPEN_BARE) != 0,
		//cache:          NewCache(),
	}
	config := repo.Config()
	loadWorkDir(repo, config, parent)
	return repo, nil
}

func loadWorkDir(repo *Repository, config *Config, parent string) {
	if repo.isBare {
		return
	}
	workTree, err := config.LookupString("core.worktree")
	if err == nil {
		path := filepath.Join(repo.pathRepository, workTree)
		repo.workDir = filepath.Clean(path)
		return
	} else if parent != "" {
		info, _ := os.Stat(parent)
		if info.IsDir() {
			repo.workDir = parent
		}
		return
	}
	repo.workDir = filepath.Dir(repo.pathRepository) + string(filepath.Separator)
}

func findRepo(startPath string, flags uint32, ceilingDirs []string) (repoPath, parentPath, linkPath string, err error) {
	path, err := filepath.Abs(startPath)
	if err != nil {
		return
	}
	ceilingDirOffset := 1
	for _, ceilingDir := range ceilingDirs {
		if strings.HasPrefix(startPath, ceilingDir) {
			offset := len(ceilingDir)
			if offset > ceilingDirOffset {
				ceilingDirOffset = offset
			}
		}
	}

	tryWithDotGit := (flags & GIT_REPOSITORY_OPEN_BARE) != 0
	if !tryWithDotGit && filepath.Base(path) != ".git" {
		path = filepath.Join(path, ".git")

	}
	for err == nil && repoPath == "" {
		var stat os.FileInfo
		stat, err = os.Stat(path)
		if err == nil {
			if stat.IsDir() {
				if isValidRepositoryPath(path) {
					repoPath = path + string(filepath.Separator)
				}
			}
			if stat.Mode().IsRegular() {
				repoLink, e := readGitFile(path)
				if e != nil {
					err = e
					return
				}
				if isValidRepositoryPath(repoLink) {
					repoPath = repoLink
					linkPath = path
				}
			}
		}
		path := filepath.Dir(path)
		if tryWithDotGit {
			if (flags&GIT_REPOSITORY_OPEN_NO_SEARCH) != 0 || len(path) <= ceilingDirOffset {
				break
			}
			path = filepath.Join(path, ".git")
		}
		tryWithDotGit = !tryWithDotGit
	}
	if err == nil && (flags&GIT_REPOSITORY_OPEN_BARE) == 0 {
		if len(repoPath) == 0 {
			parentPath = ""
		} else {
			parentPath = filepath.Dir(path) + string(filepath.Separator)
		}
	}
	if repoPath == "" && err == nil {
		err = errors.New(fmt.Sprintf("Could not find repository from '%s'", startPath))
	}
	return
}

func readGitFile(path string) (string, error) {
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(contentBytes)
	if !strings.HasPrefix(content, "gitdir:") {
		return "", errors.New(".git file shoudl have 'gitdir:' prefix")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(path), strings.TrimSpace(content[7:]))), nil
}

func isContainsFile(dir, fileName string) bool {
	stat, err := os.Stat(filepath.Join(dir, fileName))
	if err != nil {
		return false
	}
	return stat.Mode().IsRegular()
}

func isContainsDir(dir, subDirName string) bool {
	stat, err := os.Stat(filepath.Join(dir, subDirName))
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func isValidRepositoryPath(repositoryPath string) bool {
	return isContainsDir(repositoryPath, GIT_OBJECTS_DIR) &&
		isContainsFile(repositoryPath, GIT_HEAD_FILE) &&
		isContainsDir(repositoryPath, GIT_REFS_DIR)
}
