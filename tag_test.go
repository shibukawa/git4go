package git4go

import (
	"./testutil"
	"strings"
	"testing"
)

func Test_LookupTag(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo")
	oid, _ := NewOid("7b4384978d2493e851f9cca7858815fac9b10980")
	tag, err := repo.LookupTag(oid)
	if err != nil {
		t.Error("err should be nil: ", err)
	}
	if tag == nil {
		t.Error("tag should not be nil")
	} else {
		if tag.Name() != "e90810b" {
			t.Error("tag name was wrong: ", tag.Name())
		}
		if tag.TargetType() != ObjectCommit {
			t.Error("tag type was wrong:", tag.TargetType())
		}
		exactOid, _ := NewOid("e90810b8df3e80c413d903f631643c716887138d")
		if !tag.TargetId().Equal(exactOid) {
			t.Error("tag target object id was wrong: ", tag.TargetId())
		}
		if !strings.Contains(tag.Message(), "This is a very simple tag.") {
			t.Error("tag message was wrong: ", tag.Message())
		}
		if tag.Tagger().Name != "Vicent Marti" {
			t.Error("tagger was wrong: ", tag.Tagger().Name)
		}
	}
}

func Test_ListTag(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo")
	tags, err := repo.ListTag()
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if len(tags) == 0 {
		t.Error("result should contain tags:", tags)
	}
}

func Test_ListTagInPackFile(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo2")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo2")
	tags, err := repo.ListTag()
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if len(tags) != 2 {
		t.Error("result should contain tags:", tags)
	}
}

func Test_ReadTagTaggerLess(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()
	repo, _ := OpenRepository("test_resources/testrepo.git")
	oid, _ := NewOid("4a23e2e65ad4e31c4c9db7dc746650bfad082679")
	tag, err := repo.LookupTag(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	} else {
		if tag.Name() != "taggerless" {
			t.Error("wrong tag name:", tag.Name())
		}
		if tag.TargetType() != ObjectCommit {
			t.Error("wrong type", tag.TargetType())
		}
		if tag.Tagger() != nil {
			t.Error("tagger should be nil:", tag.Tagger())
		}
		if tag.Message() != "" {
			t.Error("message should be empty:", tag.Message())
		}
	}

}
