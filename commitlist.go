package git4go

import (
	"sort"
)

type CommitListFlag uint

const (
	Parent1 CommitListFlag = 1 << iota
	Parent2 CommitListFlag = 1 << iota
	Result  CommitListFlag = 1 << iota
	Stale   CommitListFlag = 1 << iota
)

type commitListNode struct {
	oid           *Oid
	time          uint64
	seen          bool
	uninteresting bool
	topologyDelay bool
	parsed        bool
	inDegree      int
	flags         CommitListFlag

	parents []*commitListNode
}

type commitListNodes []*commitListNode

func (q commitListNodes) interesting() bool {
	for _, commit := range q {
		if (commit.flags & Stale) == 0 {
			return true
		}
	}
	return false
}

func (q commitListNodes) contains(node *commitListNode) bool {
	for _, commit := range q {
		if commit == node {
			return true
		}
	}
	return false
}

func (q commitListNodes) interestingArr() bool {
	for _, n := range q {
		if !n.uninteresting {
			return true
		}
	}
	return false
}

func (q commitListNodes) Len() int {
	return len(q)
}
func (q commitListNodes) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}
func (q commitListNodes) Less(i, j int) bool {
	return q[i].time > q[j].time
}

func (q commitListNodes) insertByTime(commit *commitListNode) commitListNodes {
	result := append(q, commit)
	sort.Sort(result)
	return result
}
