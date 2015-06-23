package git4go

import (
	"./testutil"
	"strings"
	"testing"
)

func Test_LookupBlob(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/testrepo.git")
	if err != nil {
		t.Error("it should be null when loading repository in success")
	}
	if repo == nil {
		t.Error("it should load repository")
		return
	}

	oid, _ := NewOid("0266163a49e280c4f5ed1e08facd36a2bd716bcf")
	blob, err := repo.LookupBlob(oid)
	if err != nil {
		t.Error("it should be nil", err)
	}
	if blob == nil {
		t.Error("obj should not be nil")
	} else {
		if blob.Type() != ObjectBlob {
			t.Error("obj type is wrong:", blob.Type(), ObjectBlob)
		}
		size := blob.Size()
		content := blob.Contents()
		if !strings.HasPrefix(string(content), "Testing a readme.txt") {
			t.Error("invalid content")
		}
		if size == 0 {
			t.Error("size is invalid")
		}
	}
}

func Test_CreateBlobFromBuffer(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/empty_standard_repo/")
	id, err := repo.CreateBlobFromBuffer([]byte{})

	if err != nil {
		t.Error("error should be nil")
	} else {
		if id.String() != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
			t.Fatal("Empty buffer did not deliver empty blob id")
		}
	}
}
