package git4go

import (
	"./testutil"
	"bytes"
	//"fmt"
	"github.com/Unknwon/goconfig"
	"testing"
)

func Test_DefaultSignature(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	file, err := goconfig.LoadConfigFile("test_resources/empty_standard_repo/.git/config")
	file.SetValue("user", "name", "TestUser")
	file.SetValue("user", "email", "user@example.com")
	goconfig.SaveConfigFile(file, "test_resources/empty_standard_repo/.git/config")

	repo, err := OpenRepository("test_resources/empty_standard_repo")
	if err != nil {
		t.Error("it should be null when loading repository in success")
	}
	if repo == nil {
		t.Error("it should load repository")
		return
	} else {
		config := repo.Config()
		if name, _ := config.LookupString("user.name"); name != "TestUser" {
			t.Error("test setup error: user.name", name)
		}
		if email, _ := config.LookupString("user.email"); email != "user@example.com" {
			t.Error("test setup error: user.email", email)
		}
		sig, err := repo.DefaultSignature()
		if err != nil {
			t.Error("err should be nil", err)
		}
		if sig == nil || sig.Name != "TestUser" {
			t.Error("name is wrong:", sig.Name)
		}
		if sig == nil || sig.Email != "user@example.com" {
			t.Error("email is wrong:", sig.Email)
		}
	}
}

var parseSource string = `dummy
author Shawn O. Pearce <spearce@spearce.org> 1225475778 -0700
committer Shawn O. Pearce <spearce@spearce.org> 1225476305 -0700

Add a git_sobj_close to release the git_sobj data

Signed-off-by: Shawn O. Pearce <spearce@spearce.org>
`

func Test_parseSignature(t *testing.T) {
	buffer := []byte(parseSource)
	offset := len("dummy\n")
	signature, offset2, err := parseSignature(buffer, offset, []byte("author "))
	if err != nil {
		t.Error("err should be nil:", err)
		return
	}
	if signature == nil {
		t.Error("signature should not be nil")
	} else {
		if signature.Name != "Shawn O. Pearce" {
			t.Error("parse error: name", signature.Name)
		}
		if signature.Email != "spearce@spearce.org" {
			t.Error("parse error: email", signature.Email)
		}
		if signature.When.Year() != 2008 {
			t.Error("parse error: when", signature.When.String())
		}
		if signature.When.Hour() != 2 {
			t.Error("parse error: when", signature.When.String())
		}
		_, diff := signature.When.Zone()
		if diff != -(7 * 3600) {
			t.Error("parse error: time zone")
		}
	}
	if offset == offset2 {
		t.Error("offset should be different value if parse successfully")
	}
	if !bytes.Equal([]byte("committer"), buffer[offset2:offset2+len("committer")]) {
		t.Error("offset is wrong")
	}
}
