package git4go

import (
	"./testutil"
	"testing"
)

func Test_OdbHash(t *testing.T) {
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
		oid, err := odb.Hash(entry.Data, TypeString2Type(entry.Type))
		if err != nil {
			t.Error("Error should be nil: ", err, entry.Name)
		} else {
			if oid.String() != entry.Id {
				t.Error("Id should be same: ", entry.Name)
			}
		}
	}
}
