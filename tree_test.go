package git4go

import (
	"./testutil"
	"testing"
)

/*
b0a8568a7614806378a54db5706ee3b06ae58693

100644 blob fd8430bc864cfcd5f10e5590f8a447e01b942bfe	.HEADER
100644 blob a6a1a6fa11f7d0c989afae4695d4661514cda8c8	.gitignore
100644 blob 575cdc563801dcbef0ff667322c8d00176771516	CONVENTIONS
100644 blob c36f4cf1e38ec1bb9d9ad146ed572b89ecfc9f18	COPYING
100644 blob bbd29af5b49251be2a6498bd84b488bb4304ae96	Makefile
100644 blob b27d0a8066fd0fbddfcf8a30b4e77760147b0817	api.doxygen
040000 tree ce53c27f666673c2af8d406447078ea03bd95f6b	include
040000 tree 938317c9319e3280b38d705ad1cf74830cb39eff	src
040000 tree 81aba2b4fc907644c29735ebcceffa4ce01dd23a	tests
*/

func Test_LookupTree(t *testing.T) {
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

	oid, _ := NewOid("b0a8568a7614806378a54db5706ee3b06ae58693")
	tree, err := repo.LookupTree(oid)
	if err != nil {
		t.Error("it should be nil", err)
	}
	if tree == nil {
		t.Error("obj should not be nil")
	} else {
		if tree.EntryCount() != 9 {
			t.Error("tree should have childs")
		}
		entry := tree.EntryByIndex(2)
		if entry == nil || entry.Name != "CONVENTIONS" || entry.Filemode != FilemodeBlob || entry.Type != ObjectBlob {
			t.Error("entry CONVENTIONS is invalid", entry)
		}
		entry = tree.EntryByName("src")
		if entry == nil || entry.Name != "src" || entry.Filemode != FilemodeTree || entry.Type != ObjectTree {
			t.Error("entry src is invalid", entry)
		}
	}
}
