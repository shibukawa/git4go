package git4go

import (
	"./testutil"
	"testing"
)

func Test_LooseExists_Success(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	testutil.One.Write()

	odb, _ := OdbOpen("test-objects")
	id, err := NewOid(testutil.One.Id)
	if err != nil {
		t.Error("id parse error:", err)
	}

	if !odb.Exists(id) {
		t.Error("id should be exists")
	}
}

func Test_LooseExists_Failure(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	testutil.One.Write()

	odb, _ := OdbOpen("test-objects")
	noExistsId, err := NewOid("8b137891791fe96927ad78e64b0aad7bded08baa")
	if err != nil {
		t.Error("id parse error:", err)
	}

	if odb.Exists(noExistsId) {
		t.Error("id should not be exists")
	}
}

func Test_LooseExistsPrefix_Success(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	testutil.One.Write()

	odb, _ := OdbOpen("test-objects")
	id, err := NewOidFromPrefix(testutil.One.Id[:8])
	if err != nil {
		t.Error("short id parse error:", err)
	}
	id2, err := odb.ExistsPrefix(id, 8)

	if id2 == nil {
		t.Error("id should be exists")
	} else if id2.String() != testutil.One.Id {
		t.Error("id should be same")
	}
	if err != nil {
		t.Error("err should be nil:", err)
	}
}

func Test_LooseExistsPrefix_Failure(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	testutil.One.Write()

	odb, _ := OdbOpen("test-objects")
	id, err := NewOidFromPrefix("8b13789a")
	if err != nil {
		t.Error("short id parse error:", err)
	}
	id2, err := odb.ExistsPrefix(id, 8)

	if id2 != nil {
		t.Error("id should be nil")
	}
	if err == nil {
		t.Error("err should be nil:", err)
	}
}
