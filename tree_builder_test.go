package git4go

import (
	"./testutil"
	"testing"
)

/*
$ git ls-tree b10489944b9ead17427551759d180d10203e06ba                                                                                                                                                [master]
100644 blob 1a039633309bdb88eb5e6c46d1f8c2ade51f09e6	commit.c
100644 blob cb98fecad0c4379496f32a3f7ce3aea2ca553b3a	odb.c
100644 blob f15b75039d7d4db8e989b85fc085e6f1412ea318	oid.c
100644 blob 4801a2389fa38676d2a42061f5e3841324667b10	revwalk.c
*/

func Test_TreeBuilder(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()
	repo, _ := OpenRepository("test_resources/empty_standard_repo/.git")

	builder, _ := repo.TreeBuilder()

	oid1, _ := NewOid("1a039633309bdb88eb5e6c46d1f8c2ade51f09e6")
	builder.Insert("commit.c", oid1, 0100644)
	oid2, _ := NewOid("cb98fecad0c4379496f32a3f7ce3aea2ca553b3a")
	builder.Insert("odb.c", oid2, 0100644)
	oid3, _ := NewOid("f15b75039d7d4db8e989b85fc085e6f1412ea318")
	builder.Insert("oid.c", oid3, 0100644)
	oid4, _ := NewOid("4801a2389fa38676d2a42061f5e3841324667b10")
	builder.Insert("revwalk.c", oid4, 0100644)

	if len(builder.Entries) != 4 {
		t.Error("insert should work")
	} else {
		oid, err := builder.Write()
		if err != nil {
			t.Error("error should be nil")
		} else if oid == nil {
			t.Error("oid should not be nil")
		} else {
			correctOid, _ := NewOid("b10489944b9ead17427551759d180d10203e06ba")
			if !correctOid.Equal(oid) {
				t.Error("resulting oid should become correct oid:", oid.String())
			}
		}
	}
}
