package git4go

import (
	"testing"
)

func TestFnMatch_NoMetaCharacters(t *testing.T) {
	if !fnMatch("test", "test", 0) {
		t.Error("match error")
	}
	if fnMatch("git", "mercurial", 0) {
		t.Error("match error")
	}
	if fnMatch("git", "git2", 0) {
		t.Error("match error")
	}
	if fnMatch("git2", "git", 0) {
		t.Error("match error")
	}
	if fnMatch("git2", "git", 0) {
		t.Error("match error")
	}
	// cases
	if !fnMatch("teSt", "tEst", FNMCaseFold) {
		t.Error("match error")
	}
	if fnMatch("teSt", "tEst2", FNMCaseFold) {
		t.Error("match error")
	}
}

func TestFnMatch_Escape(t *testing.T) {
	// escape
	if !fnMatch("\\test", "\\test", 0) {
		t.Error("match error")
	}
	if fnMatch("\\test", "test", 0) {
		t.Error("match error")
	}
	if fnMatch("\\test", "\\test2", 0) {
		t.Error("match error")
	}

	// no escape
	if !fnMatch("\\test", "\\test", FNMNoEscape) {
		t.Error("match error")
	}
	if !fnMatch("\\test", "\\test", FNMNoEscape) {
		t.Error("match error")
	}
	if fnMatch("\\test", "\\test2", FNMNoEscape) {
		t.Error("match error")
	}
}

func TestFnMatch_Star(t *testing.T) {
	// no path mode
	if !fnMatch("refs/*", "refs/heads/master", 0) {
		t.Error("match error")
	}

	// path mode
	if fnMatch("refs/*", "refs/heads/master", FNMPathName) {
		t.Error("match error")
	}
	if !fnMatch("refs/heads/*", "refs/heads/master", FNMPathName) {
		t.Error("match error")
	}
	if !fnMatch("refs/*/master", "refs/heads/master", FNMPathName) {
		t.Error("match error")
	}
	if fnMatch("refs/*/awesome", "refs/heads/feature/awesome", FNMPathName) {
		t.Error("match error")
	}
	// todo: bug?
	//if !fnMatch("refs/**/awesome", "refs/heads/feature/awesome", FNMPathName) {
	//	t.Error("match error")
	//}
}
