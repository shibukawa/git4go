package git4go

import (
	"./testutil"
	"testing"
)

/*
111d5ccf0bb010c4e8d7af3eedfa12ef4c5e265b

tree 50330c02bd4fd95c9db1fcf2f97f4218e42b7226
parent b51eb250ed0cbda59d3108d04569fab9413909fd
author Shawn O. Pearce <spearce@spearce.org> 1225475778 -0700
committer Shawn O. Pearce <spearce@spearce.org> 1225476305 -0700

Add a git_sobj_close to release the git_sobj data

Signed-off-by: Shawn O. Pearce <spearce@spearce.org>
*/

func Test_LookupCommit(t *testing.T) {
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

	oid, _ := NewOid("111d5ccf0bb010c4e8d7af3eedfa12ef4c5e265b")
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		t.Error("it should be nil", err)
	}
	if commit == nil {
		t.Error("obj should not be nil")
	} else {
		tree, err := commit.Tree()
		if err != nil {
			t.Error("err should be nil:", err)
		} else {
			if !tree.Id().Equal(commit.treeId) {
				t.Error("tree should have same Oid")
			}
		}
	}
}
