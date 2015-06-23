package git4go

import (
	"./testutil"
	"testing"
)

func Test_TypeString2Type(t *testing.T) {
	if TypeString2Type("blob") != ObjectBlob {
		t.Error("blob type convert error")
	}
	if TypeString2Type("commit") != ObjectCommit {
		t.Error("commit type convert error")
	}
	if TypeString2Type("tree") != ObjectTree {
		t.Error("tree type convert error")
	}
	if TypeString2Type("tag") != ObjectTag {
		t.Error("tag type convert error")
	}
	if TypeString2Type("CoMmIt") != ObjectBad {
		t.Error("invalid type convert error")
	}
	if TypeString2Type("") != ObjectBad {
		t.Error("invalid type convert error2")
	}
}

func Test_Type2TypeString(t *testing.T) {
	if ObjectBlob.String() != "blob" {
		t.Error("blob type convert error")
	}
	if ObjectCommit.String() != "commit" {
		t.Error("commit type convert error")
	}
	if ObjectTree.String() != "tree" {
		t.Error("tree type convert error")
	}
	if ObjectTag.String() != "tag" {
		t.Error("tag type convert error")
	}
	if ObjectBad.String() != "" {
		t.Error("invalid type convert error")
	}
}

func Test_LookupObject(t *testing.T) {
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

	objects := map[string]ObjectType{
		"0266163a49e280c4f5ed1e08facd36a2bd716bcf": ObjectBlob,
		//"53fc32d17276939fc79ed05badaef2db09990016": ObjectTree,
		//"6dcf9bf7541ee10456529833502442f385010c3d": ObjectCommit,
	}
	for oidString, objType := range objects {
		oid, _ := NewOid(oidString)
		obj, err := repo.Lookup(oid)
		if err != nil {
			t.Error("it should be nil", err)
		}
		if obj == nil {
			t.Error("obj should not be nil")
		} else if obj.Type() != objType {
			t.Error("obj type is wrong:", obj.Type(), objType)
		}
	}
}
