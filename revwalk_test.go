package git4go

import (
	"./testutil"
	"io/ioutil"
	"testing"
)

/*
   *   a4a7dce [0] Merge branch 'master' into br2
   |\
   | * 9fd738e [1] a fourth commit
   | * 4a202b3 [2] a third commit
   * | c47800c [3] branch commit one
   |/
   * 5b5b025 [5] another commit
   * 8496071 [4] testing
*/

var commitHead string = "a4a7dce85cf63874e984719f4fdd239f5145052f"
var commitIds []string = []string{
	"a4a7dce85cf63874e984719f4fdd239f5145052f", /* 0 */
	"9fd738e8f7967c078dceed8190330fc8648ee56a", /* 1 */
	"4a202b346bb0fb0db7eff3cffeb3c70babbd2045", /* 2 */
	"c47800c7266a2be04c571c04d5a6614691ea99bd", /* 3 */
	"8496071c1b46c854b31185ea97743be6a8774479", /* 4 */
	"5b5b025afb0b4c913b4c338a42934a3863bf3644", /* 5 */
}

var commitSortingTopology [][]int = [][]int{
	{0, 1, 2, 3, 5, 4}, {0, 3, 1, 2, 5, 4},
}

var commitSortingTime [][]int = [][]int{
	{0, 3, 1, 2, 5, 4},
}

var commitSortingTopologyReverse [][]int = [][]int{
	{4, 5, 3, 2, 1, 0}, {4, 5, 2, 1, 3, 0},
}

var commitSortingTimeReverse [][]int = [][]int{
	{4, 5, 2, 1, 3, 0},
}

var commitSortingSegment [][]int = [][]int{
	{1, 2, -1, -1, -1, -1},
}

func getCommitIndex(oid *Oid) int {
	oidString := oid.String()
	for i, commitId := range commitIds {
		if oidString == commitId {
			return i
		}
	}
	return -1
}

func checkWalkOnly(walk *RevWalk, possibleResults [][]int, t *testing.T) bool {
	resultArray := []int{-1, -1, -1, -1, -1, -1}
	oid := new(Oid)
	i := 0
	for walk.Next(oid) == nil {
		t.Log("    oid: ", i, oid.String())
		resultArray[i] = getCommitIndex(oid)
		i++
	}
	for _, possibleResult := range possibleResults {
		match := true
		for i, value := range possibleResult {
			if resultArray[i] != value {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	t.Log("    result: ", resultArray, "  expecteds:", possibleResults)
	return false
}

func checkWalk(walk *RevWalk, root *Oid, flag SortType, possibleResults [][]int, t *testing.T) bool {
	walk.Sorting(flag)
	walk.Push(root)
	return checkWalkOnly(walk, possibleResults, t)
}

func Test_RevWalk_Basic_SortingModes(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()

	oid, _ := NewOid(commitHead)

	if !checkWalk(walk, oid, SortTime, commitSortingTime, t) {
		t.Error("sort result error (1)")
	}
	if !checkWalk(walk, oid, SortTopological, commitSortingTopology, t) {
		t.Error("sort result error (2)")
	}
	if !checkWalk(walk, oid, SortTime|SortReverse, commitSortingTimeReverse, t) {
		t.Error("sort result error (3)")
	}
	if !checkWalk(walk, oid, SortTopological|SortReverse, commitSortingTopologyReverse, t) {
		t.Error("sort result error (4)")
	}
}

func Test_RevWalk_Basic_GlobHeads(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushGlob("heads")
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 14 {
		t.Error("object count is wrong", i)
	}
}

func Test_RevWalk_Basic_GlobHeadsWithInvalid(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo")
	defer testutil.CleanupWorkspace()
	ioutil.WriteFile("test_resources/testrepo/.git/refs/heads/garbage", []byte("not-a-ref"), 0777)

	repo, _ := OpenRepository("test_resources/testrepo")
	walk, _ := repo.Walk()
	walk.PushGlob("heads")
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		t.Log(i, oid)
		i++
	}
	if i != 18 {
		t.Error("object count is wrong:", i)
	}
}

func Test_RevWalk_Basic_PushHead(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushHead()
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 7 {
		t.Error("object count is wrong", i)
	}
}

func Test_RevWalk_Basic_PushHeadHideRef(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushHead()
	walk.HideRef("refs/heads/packed-test")
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 4 {
		t.Error("object count is wrong", i)
	}
}

func Test_RevWalk_Basic_PushHeadHideRefNoBase(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushHead()
	walk.HideRef("refs/heads/packed")
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 7 {
		t.Error("object count is wrong", i)
	}
}

/*
* $ git rev-list HEAD 5b5b02 ^refs/heads/packed-test
* a65fedf39aefe402d3bb6e24df4d4f5fe4547750
* be3563ae3f795b2b4353bcce3a527ad0a4f7f644
* c47800c7266a2be04c571c04d5a6614691ea99bd
* 9fd738e8f7967c078dceed8190330fc8648ee56a

* $ git log HEAD 5b5b02 --oneline --not refs/heads/packed-test | wc -l => 4
* a65fedf
* be3563a Merge branch 'br2'
* c47800c branch commit one
* 9fd738e a fourth commit
 */

func Test_RevWalk_Basic_MultiplePush_1(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushHead()
	walk.HideRef("refs/heads/packed-test")
	id, _ := NewOid("5b5b025afb0b4c913b4c338a42934a3863bf3644")
	walk.Push(id)
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 4 {
		t.Error("object count is wrong", i)
	}
}

/*
* Difference between test_revwalk_basic__multiple_push_1 and
* test_revwalk_basic__multiple_push_2 is in the order reference
* refs/heads/packed-test and commit 5b5b02 are pushed.
* revwalk should return same commits in both the tests.

* $ git rev-list 5b5b02 HEAD ^refs/heads/packed-test
* a65fedf39aefe402d3bb6e24df4d4f5fe4547750
* be3563ae3f795b2b4353bcce3a527ad0a4f7f644
* c47800c7266a2be04c571c04d5a6614691ea99bd
* 9fd738e8f7967c078dceed8190330fc8648ee56a

* $ git log 5b5b02 HEAD --oneline --not refs/heads/packed-test | wc -l => 4
* a65fedf
* be3563a Merge branch 'br2'
* c47800c branch commit one
* 9fd738e a fourth commit
 */
func Test_RevWalk_Basic_MultiplePush_2(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.HideRef("refs/heads/packed-test")
	id, _ := NewOid("5b5b025afb0b4c913b4c338a42934a3863bf3644")
	walk.Push(id)
	walk.PushHead()
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 4 {
		t.Error("object count is wrong", i)
	}
}

func Test_RevWalk_Basic_DisallowNonCommit(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	id, _ := NewOid("521d87c1ec3aef9824daf6d96cc0ae3710766d91")
	err := walk.Push(id)
	if err == nil {
		t.Error("err should not be nil")
	}
}

func Test_RevWalk_Basic_HideThenPush(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	id, _ := NewOid("5b5b025afb0b4c913b4c338a42934a3863bf3644")
	walk.Hide(id)
	walk.Push(id)
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 0 {
		t.Error("object count is wrong", i)
	}
}

func Test_RevWalk_Basic_PushAll(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.PushGlob("*")
	i := 0
	oid := new(Oid)
	for walk.Next(oid) == nil {
		i++
	}
	if i != 15 {
		t.Error("object count is wrong", i)
	}
}

/*
* $ git rev-list br2 master e908
* a65fedf39aefe402d3bb6e24df4d4f5fe4547750
* e90810b8df3e80c413d903f631643c716887138d
* 6dcf9bf7541ee10456529833502442f385010c3d
* a4a7dce85cf63874e984719f4fdd239f5145052f
* be3563ae3f795b2b4353bcce3a527ad0a4f7f644
* c47800c7266a2be04c571c04d5a6614691ea99bd
* 9fd738e8f7967c078dceed8190330fc8648ee56a
* 4a202b346bb0fb0db7eff3cffeb3c70babbd2045
* 5b5b025afb0b4c913b4c338a42934a3863bf3644
* 8496071c1b46c854b31185ea97743be6a8774479
 */
func Test_RevWalk_MimicGitRevList(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/testrepo.git")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/testrepo.git")
	walk, _ := repo.Walk()
	walk.Sorting(SortTime)
	walk.PushRef("refs/heads/br2")
	walk.PushRef("refs/heads/master")
	refOid, _ := NewOid("e90810b8df3e80c413d903f631643c716887138d")
	walk.Push(refOid)

	oid := new(Oid)
	err := walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "a65fedf39aefe402d3bb6e24df4d4f5fe4547750" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "e90810b8df3e80c413d903f631643c716887138d" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "6dcf9bf7541ee10456529833502442f385010c3d" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "a4a7dce85cf63874e984719f4fdd239f5145052f" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "be3563ae3f795b2b4353bcce3a527ad0a4f7f644" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "c47800c7266a2be04c571c04d5a6614691ea99bd" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "9fd738e8f7967c078dceed8190330fc8648ee56a" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "4a202b346bb0fb0db7eff3cffeb3c70babbd2045" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "5b5b025afb0b4c913b4c338a42934a3863bf3644" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if err != nil {
		t.Error("err should be nil:", err)
	}
	if oid.String() != "8496071c1b46c854b31185ea97743be6a8774479" {
		t.Error("id is wrong:", oid.String())
	}
	err = walk.Next(oid)
	if !IsErrorCode(err, ErrIterOver) {
		t.Error("error code is wrong")
	}
}
