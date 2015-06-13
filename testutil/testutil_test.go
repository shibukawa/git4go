package testutil

import (
	"testing"
	"strings"
	"io/ioutil"
)

func Test_Fixture_doNothing(t *testing.T) {
	e := PrepareFixture("testdata/testfolder")
	if e != nil {
		t.Errorf("e should be nil")
	}
	e2 := CleanupFixture()
	if e2 != nil {
		t.Errorf("e2 should be nil")
	}
}

func Test_Fixture_modifyContent(t *testing.T) {
	originalContent, _ := ioutil.ReadFile("testdata/testfolder/test.txt")
	e := PrepareFixture("testdata/testfolder")
	if e != nil {
		t.Errorf("e should be nil")
	}
	ioutil.WriteFile("testdata/testfolder/test.txt", []byte("new content"), 0777)
	e2 := CleanupFixture()
	if e2 != nil {
		t.Errorf("e2 should be nil")
	}
	newContent, _ := ioutil.ReadFile("testdata/testfolder/test.txt")
	if string(originalContent) != string(newContent) {
		t.Errorf("cleanupFixture should revert changes")
	}
}

func Test_Fixture_initializeTwice(t *testing.T) {
	PrepareFixture("testdata/testfolder")
	defer CleanupFixture()
	e := PrepareFixture("testdata/testfolder")
	if e == nil {
		t.Errorf("e should not be nil")
	}
}

func Test_Fixture_cleanupTwice(t *testing.T) {
	PrepareFixture("testdata/testfolder")
	e2 := CleanupFixture()
	if e2 != nil {
		t.Errorf("e2 should be nil")
	}
	e3 := CleanupFixture()
	if e3 == nil {
		t.Errorf("e3 should not be nil")
	}
}

func Test_Workspace_doNothing(t *testing.T) {
	e := PrepareWorkspace("testdata/repo")
	defer CleanupWorkspace()
	if e != nil {
		t.Errorf("e should be nil")
	}
	HEAD, err := ioutil.ReadFile("testdata/repo/.git/HEAD")
	if err != nil {
		t.Errorf("workspace should have .git")
	}
	if !strings.HasPrefix(string(HEAD), "ref: refs/heads/master") {
		t.Errorf("it should read correct content")
	}
}

func Test_Workspace_withSubmodules(t *testing.T) {
	e := PrepareWorkspace("testdata/submodules")
	defer CleanupWorkspace()
	if e != nil {
		t.Error("e should be nil", e)
	}
	submodules, err := ioutil.ReadFile("testdata/submodules/.gitmodules")
	if err != nil {
		t.Errorf("workspace should have .gitmodules")
	}
	if !strings.HasPrefix(string(submodules), "[submodule \"testrepo\"]") {
		t.Errorf("it should read correct content of .gitmodules")
	}
	HEAD, err := ioutil.ReadFile("testdata/submodules/testrepo/.git/HEAD")
	if err != nil {
		t.Errorf("workspace's submodule should have .git")
	}
	if !strings.HasPrefix(string(HEAD), "ref: refs/heads/master") {
		t.Errorf("it should read correct content in submodule's .git")
	}
}