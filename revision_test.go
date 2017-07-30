package git4go

/*
import (
	"./testutil"
	"testing"
)

func checkObjectAndRefInRepo(spec, expectedOid, expectedRefName string, repo *Repository, t *testing.T) {
	obj, ref, err := repo.RevparseExt(spec)
	oid, err := NewOid(expectedOid)
	if err != nil {
		t.Error("id was wrong:", expectedOid)
	} else if !obj.Id().Equal(oid) {
		t.Error("Ids are not equal:", expectedOid, obj.Id().String())
	}
	if ref.Name() != expectedRefName {
		t.Error("Ref.Name() was wrong:", expectedRefName, ref.Name())
	}
}

func checkObjectInRepo(spec, expectedOid string, repo *Repository, t *testing.T) {
	obj, _, err := repo.RevparseExt(spec)
	if obj == nil {
		t.Error("obj should not be nil:", err)
	} else if expectedOid != "" {
		oid, err := NewOid(expectedOid)
		if err != nil {
			t.Error("err should be nil:", err)
		} else if !obj.Id().Equal(oid) {
			t.Error("Ids are not equal:", expectedOid, obj.Id().String())
		}
	} else if err == nil {
		t.Error("err should be error", spec)
	}
}

func checkIdInRepo(spec, expectedLeft, expectedRight string, flag RevparseFlag, repo *Repository, t *testing.T) {
	revSpec, err := repo.Revparse(spec)

	if expectedLeft != "" {
		oid, _ := NewOid(expectedLeft)
		if !revSpec.From().Id().Equal(oid) {
			t.Error("Ids are not equal(From):", expectedLeft, revSpec.From().Id().String())
		}
	} else if err == nil {
		t.Error("err should not be nil:", expectedLeft)
	}

	if expectedRight != "" {
		oid, _ := NewOid(expectedRight)
		if !revSpec.To().Id().Equal(oid) {
			t.Error("Ids are not equal(To):", expectedLeft, revSpec.From().Id().String())
		}
	}

	if flag != RevparseNone && flag != revSpec.Flags() {
		t.Error("flag was wrong:", flag, revSpec.Flags())
	}
}

func checkInvalidSingleSpec(invalidSpec string, repo *Repository, t *testing.T) {
	_, err := repo.RevparseSingle(invalidSpec)
	if err != nil {
		t.Error("Id should not be found: ", invalidSpec)
	}
}

func Test_Revparse_NonExistObject(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("this-does-not-exist", "", repo, t)
	checkObjectInRepo("this-does-not-exist^1", "", repo, t)
	checkObjectInRepo("this-does-not-exist~3", "", repo, t)
}

func Test_Revparse_InvalidRefName(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkInvalidSingleSpec("this doesn't make sense", repo, t)
	checkInvalidSingleSpec("Inv@{id", repo, t)
	checkInvalidSingleSpec("", repo, t)
}

func Test_Revparse_Sha1(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("c47800c7266a2be04c571c04d5a6614691ea99bd", "c47800c7266a2be04c571c04d5a6614691ea99bd", repo, t)
	checkObjectInRepo("c47800c", "c47800c7266a2be04c571c04d5a6614691ea99bd", repo, t)
}

func Test_Revparse_Head(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("HEAD", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("HEAD^0", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("HEAD~0", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("master", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
}

func Test_Revparse_FullRefs(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("refs/heads/master", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("refs/heads/test", "e90810b8df3e80c413d903f631643c716887138d", repo, t)
	checkObjectInRepo("refs/tags/test", "b25fa35b38051e4ae45d4222e795f9df2e43f1d1", repo, t)
}

func Test_Revparse_PartialRefs(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("point_to_blob", "1385f264afb75a56a5bec74243be9b367ba4ca08", repo, t)
	checkObjectInRepo("packed-test", "4a202b346bb0fb0db7eff3cffeb3c70babbd2045", repo, t)
	checkObjectInRepo("br2", "a4a7dce85cf63874e984719f4fdd239f5145052f", repo, t)
}

func Test_Revparse_DescribeOutput(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("blah-7-gc47800c", "c47800c7266a2be04c571c04d5a6614691ea99bd", repo, t)
	checkObjectInRepo("not-good", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
}

func Test_Revparse_NthParent(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("blah-7-gc47800c", "c47800c7266a2be04c571c04d5a6614691ea99bd", repo, t)
	checkObjectInRepo("not-good", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)

	checkInvalidSingleSpec("be3563a^-1", repo, t)
	checkInvalidSingleSpec("^", repo, t)
	checkInvalidSingleSpec("be3563a^{tree}^", repo, t)
	checkInvalidSingleSpec("point_to_blob^{blob}^", repo, t)
	checkInvalidSingleSpec("this doesn't make sense^1", repo, t)

	checkObjectInRepo("be3563a^1", "9fd738e8f7967c078dceed8190330fc8648ee56a", repo, t)
	checkObjectInRepo("be3563a^", "9fd738e8f7967c078dceed8190330fc8648ee56a", repo, t)
	checkObjectInRepo("be3563a^2", "c47800c7266a2be04c571c04d5a6614691ea99bd", repo, t)
	checkObjectInRepo("be3563a^1^1", "4a202b346bb0fb0db7eff3cffeb3c70babbd2045", repo, t)
	checkObjectInRepo("be3563a^^", "4a202b346bb0fb0db7eff3cffeb3c70babbd2045", repo, t)
	checkObjectInRepo("be3563a^2^1", "5b5b025afb0b4c913b4c338a42934a3863bf3644", repo, t)
	checkObjectInRepo("be3563a^0", "be3563ae3f795b2b4353bcce3a527ad0a4f7f644", repo, t)
	checkObjectInRepo("be3563a^{commit}^", "9fd738e8f7967c078dceed8190330fc8648ee56a", repo, t)

	checkObjectInRepo("be3563a^42", "", repo, t)
}

func Test_Revparse_NotTag(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")

	checkObjectInRepo("point_to_blob^{}", "1385f264afb75a56a5bec74243be9b367ba4ca08", repo, t)
	checkObjectInRepo("wrapped_tag^{}", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("master^{}", "a65fedf39aefe402d3bb6e24df4d4f5fe4547750", repo, t)
	checkObjectInRepo("master^{tree}^{}", "944c0f6e4dfa41595e6eb3ceecdb14f50fe18162", repo, t)
	checkObjectInRepo("e90810b^{}", "e90810b8df3e80c413d903f631643c716887138d", repo, t)
	checkObjectInRepo("tags/e90810b^{}", "e90810b8df3e80c413d903f631643c716887138d", repo, t)
	checkObjectInRepo("e908^{}", "e90810b8df3e80c413d903f631643c716887138d", repo, t)
}
*/
