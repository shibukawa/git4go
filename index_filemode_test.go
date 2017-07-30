package git4go

import (
	"./testutil"
	"testing"
	"io/ioutil"
	"path/filepath"
	"os"
	"fmt"
	"math/rand"
)

func Test_IndexReadFileMode(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/filemodes")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/filemodes")
	if repo == nil {
		t.Error("it should load repository")
		return
	}
	index, _ := repo.Index()
	if index.EntryCount() != 6 {
		t.Error("wrong entry count")
		return
	}
	expectedValues := []bool{false, true, false, true, false, true}
	for i, expected := range expectedValues {
		entry, err := index.EntryByIndex(i)
		if err != nil {
			t.Error(err)
		} else {
			if expected && (entry.Mode&0100) != 0100 {
				t.Error("entry should have root-exec flag")
			}
			if !expected && (entry.Mode&0100) == 0100 {
				t.Error("entry should not have root-exec flag")
			}
		}
	}
}

func replaceFileWithMode(fileName, backup string, createMode Filemode) {
	filePath := filepath.Join("test_resources/filemodes", fileName)
	backupPath := filepath.Join("test_resources/filemodes", backup)
	os.Rename(filePath, backupPath)
	content := fmt.Sprintf("%s as %08u (%d)", fileName, int(createMode), rand.Int())
	ioutil.WriteFile(filePath, []byte(content), os.FileMode(createMode))
}

func addAndCheckMode(index *Index, fileName string, expectedMode Filemode, t *testing.T) {
	err := index.AddByPath(fileName)
	if err != nil {
		t.Error("error should be nil:", err)
		return
	}
	pos := index.Find(fileName)
	if pos == -1 {
		t.Error("can't find file:", fileName)
		return
	}
	entry, err := index.EntryByIndex(pos)
	if err != nil {
		t.Error("error should be nil:", err)
		return
	}
	if entry.Mode != expectedMode {
		t.Errorf("File mode is wrong: expected %o actual %o", expectedMode, entry.Mode)
	}
}

func Test_IndexFileModes_Untrusted(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/filemodes")
	defer testutil.CleanupWorkspace()

	repo, _ := OpenRepository("test_resources/filemodes")
	if repo == nil {
		t.Error("it should load repository")
		return
	}
	repo.Config().SetBool("core.filemode", false)

	index, _ := repo.Index()
	if (index.Caps() & IndexCapNoFilemode) == 0 {
		t.Error("index cap mode error", index.Caps())
	}
	// 1 - add 0644 over existing 0644 -> expect 0644
	replaceFileWithMode("exec_off", "exec_off.0", 0644)
	addAndCheckMode(index, "exec_off", FilemodeBlob, t)

	// 2 - add 0644 over existing 0755 -> expect 0755
	replaceFileWithMode("exec_on", "exec_on.0", 0644)
	addAndCheckMode(index, "exec_on", FilemodeBlobExecutable, t)

	// 3 - add 0755 over existing 0644 -> expect 0644
	replaceFileWithMode("exec_off", "exec_off.1", 0755)
	addAndCheckMode(index, "exec_off", FilemodeBlob, t)

	// 4 - add 0755 over existing 0755 -> expect 0755
	replaceFileWithMode("exec_on", "exec_off.1", 0755)
	addAndCheckMode(index, "exec_on", FilemodeBlobExecutable, t)

	// 5 - add new 0644 -> expect 0644
	ioutil.WriteFile("test_resources/filemodes/new_off", []byte("blah"), 0644)
	addAndCheckMode(index, "new_off", FilemodeBlob, t)

	// 6 - add new 0755 -> expect 0644 if core.filemode == false
	ioutil.WriteFile("test_resources/filemodes/new_on", []byte("blah"), 0755)
	addAndCheckMode(index, "new_on", FilemodeBlob, t)
}