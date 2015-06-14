package git4go

import (
	"testing"
)

func Test_TypeString2Type(t *testing.T) {
	if TypeString2Type("blob") != ObjectBlob {
		t.Error("blob type convert error")
	}
	if TypeString2Type("commit") != ObjectCommit {
		t.Error("commit type convert error")
	}
	if TypeString2Type("tree") != ObjectTree {
		t.Error("tree type convert error")
	}
	if TypeString2Type("tag") != ObjectTag {
		t.Error("tag type convert error")
	}
	if TypeString2Type("CoMmIt") != ObjectBad {
		t.Error("invalid type convert error")
	}
	if TypeString2Type("") != ObjectBad {
		t.Error("invalid type convert error2")
	}
}
