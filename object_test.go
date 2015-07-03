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

func assertPeel(sha string, requestedType ObjectType, expectedSha string, expectedType ObjectType, repo *Repository, t *testing.T) {
	oid, _ := NewOid(sha)
	obj, err := repo.Lookup(oid)
	if err != nil {
		t.Error("error should be nil:", err)
	}
	peeled, err := obj.Peel(requestedType)
	if err != nil {
		t.Error("error should be nil:", err)
	}

	expectedOid, _ := NewOid(expectedSha)
	if !expectedOid.Equal(peeled.Id()) {
		t.Error("Ids are not same: ", peeled.Id(), expectedSha)
	}
	if peeled.Type() != expectedType {
		t.Error("Types are not same: ", peeled.Type(), expectedType)
	}
}

func assertPeelError(sha string, requestedType ObjectType, repo *Repository, t *testing.T) {
	oid, _ := NewOid(sha)
	obj, err := repo.Lookup(oid)
	if err != nil {
		t.Error("error should be nil:", err)
	}
	_, err = obj.Peel(requestedType)
	if err == nil {
		t.Error("error should not be nil:", err)
	}

}

func Test_Peel_SameType(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	assertPeel("e90810b8df3e80c413d903f631643c716887138d", ObjectCommit,
		"e90810b8df3e80c413d903f631643c716887138d", ObjectCommit, repo, t)
	assertPeel("7b4384978d2493e851f9cca7858815fac9b10980", ObjectTag,
		"7b4384978d2493e851f9cca7858815fac9b10980", ObjectTag, repo, t)
	assertPeel("53fc32d17276939fc79ed05badaef2db09990016", ObjectTree,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
	assertPeel("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectBlob,
		"0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectBlob, repo, t)
}

func Test_Peel_Tag(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	assertPeel("7b4384978d2493e851f9cca7858815fac9b10980", ObjectCommit,
		"e90810b8df3e80c413d903f631643c716887138d", ObjectCommit, repo, t)
	assertPeel("7b4384978d2493e851f9cca7858815fac9b10980", ObjectTree,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
	assertPeelError("7b4384978d2493e851f9cca7858815fac9b10980", ObjectBlob, repo, t)
	assertPeel("7b4384978d2493e851f9cca7858815fac9b10980", ObjectAny,
		"e90810b8df3e80c413d903f631643c716887138d", ObjectCommit, repo, t)
}

func Test_Peel_Commit(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	assertPeelError("e90810b8df3e80c413d903f631643c716887138d", ObjectBlob, repo, t)
	assertPeel("e90810b8df3e80c413d903f631643c716887138d", ObjectTree,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
	assertPeel("e90810b8df3e80c413d903f631643c716887138d", ObjectCommit,
		"e90810b8df3e80c413d903f631643c716887138d", ObjectCommit, repo, t)
	assertPeelError("e90810b8df3e80c413d903f631643c716887138d", ObjectTag, repo, t)
	assertPeel("e90810b8df3e80c413d903f631643c716887138d", ObjectAny,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
}

func Test_Peel_Tree(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	assertPeelError("53fc32d17276939fc79ed05badaef2db09990016", ObjectBlob, repo, t)
	assertPeel("53fc32d17276939fc79ed05badaef2db09990016", ObjectTree,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
	assertPeelError("53fc32d17276939fc79ed05badaef2db09990016", ObjectCommit, repo, t)
	assertPeelError("53fc32d17276939fc79ed05badaef2db09990016", ObjectTag, repo, t)
	assertPeelError("53fc32d17276939fc79ed05badaef2db09990016", ObjectAny, repo, t)
}

func Test_Peel_Blob(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	assertPeel("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectBlob,
		"0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectBlob, repo, t)
	assertPeelError("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectTree, repo, t)
	assertPeelError("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectCommit, repo, t)
	assertPeelError("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectTag, repo, t)
	assertPeelError("0266163a49e280c4f5ed1e08facd36a2bd716bcf", ObjectAny, repo, t)
}

func Test_Peel_Any(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	/* tag to commit */
	assertPeel("7b4384978d2493e851f9cca7858815fac9b10980", ObjectAny,
		"e90810b8df3e80c413d903f631643c716887138d", ObjectCommit, repo, t)

	/* commit to tree */
	assertPeel("e90810b8df3e80c413d903f631643c716887138d", ObjectAny,
		"53fc32d17276939fc79ed05badaef2db09990016", ObjectTree, repo, t)
}
