package git4go

import (
	"./testutil"
	"testing"
)

func checkConflict(entry *IndexEntry, expectedName, expectedOid string, t *testing.T) {
	if entry.Path != expectedName {
		t.Error("wrong path. expected:", expectedName, "actual:", entry.Path)
	}
	if entry.Id.String() != expectedOid {
		t.Error("wrong id. expected:", expectedOid, "actual:", entry.Id.String())
	}
}

func Test_ReadIndex(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/mergedrepo")
	defer testutil.CleanupWorkspace()

	repo, err := OpenRepository("test_resources/mergedrepo")
	if err != nil {
		t.Error("it should be null when loading repository in success")
	}
	if repo == nil {
		t.Error("it should load repository")
		return
	}
	index, err := repo.Index()
	if err != nil {
		t.Error("it should be null when loading index in success")
	}
	if index == nil {
		t.Error("index should not be nil")
		return
	}
	if index.EntryCount() != 8 {
		t.Error("entry count should be 8, but", index.EntryCount())
	}
}

func Test_IndexGetConflict(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/mergedrepo")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/mergedrepo")
	if repo == nil {
		t.Error("it should load repository")
		return
	}
	index, _ := repo.Index()
	conflict, err := index.GetConflict("conflicts-one.txt")
	if err != nil {
		t.Error("it should be nil")
	}
	checkConflict(conflict.Ancestor, "conflicts-one.txt", "1f85ca51b8e0aac893a621b61a9c2661d6aa6d81", t)
	checkConflict(conflict.Our, "conflicts-one.txt", "6aea5f295304c36144ad6e9247a291b7f8112399", t)
	checkConflict(conflict.Their, "conflicts-one.txt", "516bd85f78061e09ccc714561d7b504672cb52da", t)

	conflict2, err := index.GetConflict("conflicts-two.txt")
	if err != nil {
		t.Error("it should be nil")
	}
	checkConflict(conflict2.Ancestor, "conflicts-two.txt", "84af62840be1b1c47b778a8a249f3ff45155038c", t)
	checkConflict(conflict2.Our, "conflicts-two.txt", "8b3f43d2402825c200f835ca1762413e386fd0b2", t)
	checkConflict(conflict2.Their, "conflicts-two.txt", "220bd62631c8cf7a83ef39c6b94595f00517211e", t)
}

func Test_IndexConflictIterator(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/mergedrepo")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/mergedrepo")
	if repo == nil {
		t.Error("it should load repository")
		return
	}
	index, _ := repo.Index()
	iter, err := index.ConflictIterator()
	if err != nil {
		t.Error("it should be nil")
	}
	conflict, err := iter.Next()
	if err != nil {
		t.Error("it should be nil")
	}
	checkConflict(conflict.Ancestor, "conflicts-one.txt", "1f85ca51b8e0aac893a621b61a9c2661d6aa6d81", t)
	checkConflict(conflict.Our, "conflicts-one.txt", "6aea5f295304c36144ad6e9247a291b7f8112399", t)
	checkConflict(conflict.Their, "conflicts-one.txt", "516bd85f78061e09ccc714561d7b504672cb52da", t)

	conflict2, err := iter.Next()
	if err != nil {
		t.Error("it should be nil")
	}
	checkConflict(conflict2.Ancestor, "conflicts-two.txt", "84af62840be1b1c47b778a8a249f3ff45155038c", t)
	checkConflict(conflict2.Our, "conflicts-two.txt", "8b3f43d2402825c200f835ca1762413e386fd0b2", t)
	checkConflict(conflict2.Their, "conflicts-two.txt", "220bd62631c8cf7a83ef39c6b94595f00517211e", t)

	conflict, err = iter.Next()

	if !IsErrorCode(err, ErrIterOver) {
		t.Error("it should be iter over")
	}
	if conflict.Ancestor != nil {
		t.Error("it should be nil")
	}
	if conflict.Our != nil {
		t.Error("it should be nil")
	}
	if conflict.Their != nil {
		t.Error("it should be nil")
	}
}
