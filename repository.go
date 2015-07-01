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
	GitObjectsDir                 string = "objects/"
	GitHeadFile                   string = "HEAD"
	GitRefsDir                    string = "refs/"
	GitRefsTagsDir                string = "refs/tags"
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

func (r *Repository) Workdir() string {
	if r.isBare {
		return ""
	}
	return r.workDir
}

func (r *Repository) IsBare() bool {
	return r.isBare
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
		//cache:          NewCache(),
	}
	config := repo.Config()
	loadConfigData(repo, config)
	loadWorkDir(repo, config, parent)
	return repo, nil
}

func loadConfigData(repo *Repository, config *Config) {
	isBare, err := config.LookupBool("core.bare")
	if err == nil {
		repo.isBare = isBare
	}
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
	for repoPath == "" {
		stat, tempErr := os.Stat(path)
		if tempErr == nil {
			if stat.IsDir() {
				if isValidRepositoryPath(path) {
					repoPath = path + string(filepath.Separator)
				}
			}
			if stat.Mode().IsRegular() {
				repoLink, tempErr2 := readGitFile(path)
				if tempErr2 == nil && isValidRepositoryPath(repoLink) {
					repoPath = repoLink
					linkPath = path
				}
			}
		}
		parentDir := filepath.Dir(path)
		if !tryWithDotGit && parentDir == path {
			break
		}
		path = parentDir
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
			parentPath = filepath.Dir(repoPath[:len(repoPath)-1]) + string(filepath.Separator)
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
	return isContainsDir(repositoryPath, GitObjectsDir) &&
		isContainsFile(repositoryPath, GitHeadFile) &&
		isContainsDir(repositoryPath, GitRefsDir)
}
