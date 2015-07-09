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

func Test_DwimReferenceInPackFile(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo2")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo2")
	ref, err := repo.DwimReference("v0.9")
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if ref == nil {
		t.Error("ref should not be nil")
	} else {
		if ref.Name() != "refs/tags/v0.9" {
			t.Error("name was wrong:", ref.Name())
		}
		expectOid, _ := NewOid("5b5b025afb0b4c913b4c338a42934a3863bf3644")
		if !ref.Target().Equal(expectOid) {

		}
	}
}

func Test_ReferenceResolve(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/testrepo/")

	if err != nil {
		t.Error("it should be null when loading repository in success")
	}

	if repo == nil {
		t.Error("it should load repository")
	} else {
		ref, err := repo.LookupReference("HEAD")

		if err != nil {
			t.Error("err should be null", err)
		}
		if ref == nil {
			t.Error("ref should not be null")
		} else {
			resolved, err := ref.Resolve()
			if err != nil {
				t.Error("err should be null", err)
			} else {
				if resolved.Target() == nil {
					t.Error("resolved ref should have target Oid")
				} else if resolved.Target().String() != "099fabac3a9ea935598528c27f866e34089c2eff" {
					t.Error("resolved ref should be correct hex: ", ref.Target().String())
				}
			}
		}
	}
}

func Test_ForEachReferenceName(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo/")
	var names []string
	err := repo.ForEachReferenceName(func(name string) error {
		names = append(names, name)
		return nil
	})
	if err != nil {
		t.Error("err should be nil", err)
	}
	if len(names) != 15 {
		t.Error("it should have references in repository:", len(names), names)
	}
}

func Test_ForEachReference(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo/")
	var names []string
	err := repo.ForEachReference(func(ref *Reference) error {
		names = append(names, ref.Name())
		return nil
	})
	if err != nil {
		t.Error("err should be nil", err)
	}
	if len(names) != 15 {
		t.Error("it should have references in repository:", len(names), names)
	}
}

func Test_ForEachGlobReference(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo/")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo/")
	var names []string
	err := repo.ForEachGlobReferenceName("refs/tags/*", func(name string) error {
		names = append(names, name)
		return nil
	})
	if err != nil {
		t.Error("err should be nil")
	}
	if len(names) != 6 {
		t.Error("it should have references in repository:", len(names), names)
	}
}
