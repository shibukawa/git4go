package git4go

import (
	"./testutil"
	"os"
	"strings"
	"testing"
)

// internal functions
func Test_isValidRepositoryPath(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	if isValidRepositoryPath("test_resources/empty_standard_repo") {
		t.Errorf("It is invalid path because it is not initialized")
	}

	if !isValidRepositoryPath("test_resources/empty_standard_repo/.git") {
		t.Errorf("It should be valid path")
	}
}

func Test_readGitFile(t *testing.T) {
	path, err := readGitFile("test_resources/submod2/sm_unchanged/.gitted")
	if err != nil {
		t.Error("it shouldn't be error:", err)
	}
	if path != "test_resources/submod2/.git/modules/sm_unchanged" {
		t.Error("it should return actual path, but ", path)
	}
}

func Test_Discover_standardRepository(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repoPath, err := Discover("test_resources/empty_standard_repo/", false, []string{})
	if err != nil {
		t.Error(err)
	}
	if !strings.HasSuffix(repoPath, "test_resources/empty_standard_repo/.git/") {
		t.Error("result was wrong:", repoPath)
	}
}

func Test_Discover_bareRepository(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repoPath, err := Discover("test_resources/testrepo.git", false, []string{})
	if err != nil {
		t.Error(err)
	}
	if !strings.HasSuffix(repoPath, "test_resources/testrepo.git/") {
		t.Error("result was wrong:", repoPath)
	}
}

func Test_OpenRepository_success_withStandardRepo_1(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/empty_standard_repo/")

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}

	if repo == nil {
		t.Error("it should load repository")
	} else {
		if !strings.HasSuffix(repo.Path(), "test_resources/empty_standard_repo/.git/") {
			t.Errorf("it should have correct repository path: %s", repo.Path())
		}
		if !strings.HasSuffix(repo.WorkDir(), "test_resources/empty_standard_repo/") {
			t.Errorf("it should have correct workdir path: %s", repo.Path())
		}
	}
}

func Test_OpenRepository_success_withStandardRepo_2(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/empty_standard_repo/.git")
	if repo == nil {
		t.Error("it should load repository")
	}

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}
}

func Test_OpenRepository_failure_1(t *testing.T) {
	repo, err := OpenRepository("test_resources/empty_standard_repo/")
	if repo != nil {
		t.Error("it should not load repository")
	}

	if err == nil {
		t.Error("it should not be null when loading repository in failure")
	}
}

func Test_OpenRepository_failure_2(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	os.MkdirAll("test_resources/empty_standard_repo/subdir", 0777)

	repo, err := OpenRepository("test_resources/empty_standard_repo/subdir")
	if repo != nil {
		t.Error("it should not load repository")
	}

	if err == nil {
		t.Error("it should not be null when loading repository in failure")
	}
}

func Test_OpenRepositoryExtended_success_1(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepositoryExtended("test_resources/empty_standard_repo/")
	if repo == nil {
		t.Error("it should load repository")
	}

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}
}

func Test_OpenRepositoryExtended_success_2(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepositoryExtended("test_resources/empty_standard_repo/.git")
	if repo == nil {
		t.Error("it should load repository")
	}

	if err != nil {
		t.Error("it should be null when loading repository in success", err)
	}
}

func Test_OpenRepositoryExtended_success_3(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	os.MkdirAll("test_resources/empty_standard_repo/subdir", 0777)

	repo, err := OpenRepositoryExtended("test_resources/empty_standard_repo/subdir")
	if repo == nil {
		t.Error("it should load repository")
	}

	if err != nil {
		t.Error("it should be null when loading repository in success", err)
	}
}

func Test_OpenRepositoryExtended_failure(t *testing.T) {
	repo, err := OpenRepositoryExtended(os.TempDir())
	if repo != nil {
		t.Error("it should not load repository")
	}

	if err == nil {
		t.Error("it should not be null when loading repository in failure")
	}
}
