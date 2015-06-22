package git4go

import (
	"./testutil"
	"testing"
	//"fmt"
)

func Test_PackedOdb_Exists(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()
	odb, _ := OdbOpen("test_resources/testrepo.git/objects")

	for i, packedObject := range testutil.PackedObjects {
		oid, _ := NewOid(packedObject)
		if !odb.Exists(oid) {
			t.Error("Object should exist: ", i)
		}
	}
}

func Test_PackedOdb_ExistsPrefix(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()
	odb, _ := OdbOpen("test_resources/testrepo.git/objects")

	for i, packedObject := range testutil.PackedObjects {
		shortOid, _ := NewOidFromPrefix(packedObject[:8])
		oid, _ := NewOid(packedObject)
		longOid, err := odb.ExistsPrefix(shortOid, 8)
		if err != nil {
			t.Error("Object should exist: ", i, err)
		} else if !oid.Equal(longOid) {
			t.Error("Found id hould be same with original:", oid.String(), longOid.String())
		}
	}
}

func Test_PackedOdb_ReadAndReadHeader(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()
	odb, _ := OdbOpen("test_resources/testrepo.git/objects")

	for i, packedObject := range testutil.PackedObjects {
		oid, _ := NewOid(packedObject)
		obj, err := odb.Read(oid)
		if err != nil {
			t.Error("err should be nil: ", i, err)
		}
		if obj == nil {
			t.Error("Can't read object", i)
		} else {
			objType, size, err := odb.ReadHeader(oid)
			if err != nil {
				t.Error("err should be nil: ", i, err)
			}
			if obj.Type != objType {
				t.Error("type is wrong", i, obj.Type, objType, oid.String())
			}
			if uint64(len(obj.Data)) != size {
				t.Error("size is wrong", i, len(obj.Data), size, oid.String())
			}
			//fmt.Println(string(obj.Data))
		}
	}
}
