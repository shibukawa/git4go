package git4go

import (
	"./testutil"
	"bytes"
	"os"
	"path/filepath"
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

func Test_LooseRead(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	odb, _ := OdbOpen("test-objects")

	testEntries := []*testutil.ObjectData{
		&testutil.One,
		&testutil.Commit,
		&testutil.Tree,
		&testutil.Tag,
		&testutil.Zero,
		&testutil.Two,
		&testutil.Some,
	}
	for _, entry := range testEntries {
		entry.Write()

		id, _ := NewOid(entry.Id)
		odbObject, err := odb.Read(id)
		if err != nil {
			t.Error("Error should be nil: ", err, entry.Name)
		} else {
			if odbObject.Type != TypeString2Type(entry.Type) {
				t.Error("Type should be same: ", entry.Name)
			}
			if bytes.Compare(odbObject.Data, entry.Data) != 0 {
				t.Error("Data should be same: ", entry.Name)
			}
		}
	}
}

func Test_LooseReadPrefix(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	odb, _ := OdbOpen("test-objects")

	testEntries := []*testutil.ObjectData{
		&testutil.One,
		&testutil.Commit,
		&testutil.Tree,
		&testutil.Tag,
		&testutil.Zero,
		&testutil.Two,
		&testutil.Some,
	}
	for _, entry := range testEntries {
		entry.Write()

		id, _ := NewOidFromPrefix(entry.Id[:8])
		foundId, odbObject, err := odb.ReadPrefix(id, 8)
		if err != nil {
			t.Error("Error should be nil: ", err, entry.Name)
		} else {
			if foundId.String() != entry.Id {
				t.Error("Id should be same: ", entry.Name)
			}
			if odbObject.Type != TypeString2Type(entry.Type) {
				t.Error("Type should be same: ", entry.Name)
			}
			if bytes.Compare(odbObject.Data, entry.Data) != 0 {
				t.Error("Data should be same: ", entry.Name)
			}
		}
	}
}

func Test_LooseReadHeader(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	odb, _ := OdbOpen("test-objects")

	testEntries := []*testutil.ObjectData{
		&testutil.One,
		&testutil.Commit,
		&testutil.Tree,
		&testutil.Tag,
		&testutil.Zero,
		&testutil.Two,
		&testutil.Some,
	}
	for _, entry := range testEntries {
		entry.Write()

		id, _ := NewOid(entry.Id)
		objType, size, err := odb.ReadHeader(id)
		if err != nil {
			t.Error("Error should be nil: ", err, entry.Name)
		} else {
			if objType != TypeString2Type(entry.Type) {
				t.Error("Type should be same: ", entry.Name)
			}
			if size != int64(len(entry.Data)) {
				t.Error("Data size should be same: ", entry.Name)
			}
		}
	}
}

func Test_LooseWrite(t *testing.T) {
	testutil.PrepareEmptyWorkDir("test-objects")
	defer testutil.CleanupEmptyWorkDir()
	odb, _ := OdbOpen("test-objects")

	data := "Test data\n"
	oid, err := odb.Write([]byte(data), ObjectBlob)
	if err != nil {
		t.Error("write should finish successfully: ", err)
	} else {
		if oid.String() != "67b808feb36201507a77f85e6d898f0a2836e4a5" {
			t.Error("id is wrong: ", oid.String())
		}
		_, err = os.Stat(filepath.Join("test-objects", "67", "b808feb36201507a77f85e6d898f0a2836e4a5"))
		if !os.IsNotExist(err) {
			t.Error("file is missing")
		}
	}
}
