package git4go

import (
	"./testutil"
	"testing"
)

func Test_RepositoryHead(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/testrepo/")

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}

	if repo == nil {
		t.Error("it should load repository")
	} else {
		ref, err := repo.Head()

		if err != nil {
			t.Error("err should be null", err)
		}
		if ref == nil {
			t.Error("ref should not be null")
		} else if ref.Target() == nil {
			t.Error("ref.Target() should not be null", ref)
		} else {
			if ref.Target().String() != "099fabac3a9ea935598528c27f866e34089c2eff" {
				t.Error("ref should be correct hex: ", ref.Target().String())
			}
		}
	}
}

func Test_DwimReference(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/testrepo/")

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}

	if repo == nil {
		t.Error("it should load repository")
	} else {
		ref, err := repo.DwimReference("master")

		if err != nil {
			t.Error("err should be null", err)
		}
		if ref == nil {
			t.Error("ref should not be null")
		} else if ref.Target() == nil {
			t.Error("ref.Target() should not be null", ref)
		} else {
			if ref.Target().String() != "099fabac3a9ea935598528c27f866e34089c2eff" {
				t.Error("ref should be correct hex: ", ref.Target().String())
			}
		}
	}
}
