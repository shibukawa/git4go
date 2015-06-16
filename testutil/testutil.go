package testutil

import (
	"fmt"
	"os"
	"errors"
	"math/rand"
	"path/filepath"
	"io"
	"code.google.com/p/gcfg"
)

var backupWorkspacePath string
var currentWorkspacePath string

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	defer sourcefile.Close()
	if err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	defer destFile.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, sourcefile)
	if err == nil {
		sourceInfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceInfo.Mode())
		}
	}
	return nil
}

func copyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	// create dest dir
	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	defer directory.Close()
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourceFilePointer := source + "/" + obj.Name()
		destinationFilePointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourceFilePointer, destinationFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourceFilePointer, destinationFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func PrepareFixture(workspacePath string) error {
	if backupWorkspacePath != "" {
		return errors.New("Fixture is initialized")
	}
	sourceInfo, err := os.Stat(workspacePath)
	if os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Workspace name %s is invalid", workspacePath))
	}
	if !sourceInfo.IsDir() {
		return errors.New(fmt.Sprintf("Workspace %s is not directory", workspacePath))
	}
	backupWorkspacePath = filepath.Join(os.TempDir(), fmt.Sprintf("testutil_PrepareFixture_%v", rand.Uint32()))
	currentWorkspacePath = workspacePath
	return copyDir(workspacePath, backupWorkspacePath)
}

func CleanupFixture() error {
	if backupWorkspacePath == "" {
		return errors.New("Fixture is not initialized")
	}
	os.RemoveAll(currentWorkspacePath)
	copyDir(backupWorkspacePath, currentWorkspacePath)
	os.RemoveAll(backupWorkspacePath)
	backupWorkspacePath = ""
	currentWorkspacePath = ""
	return nil
}

func PrepareWorkspace(workspacePath string) error {
	err := PrepareFixture(workspacePath)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(workspacePath, ".gitted"), filepath.Join(workspacePath, ".git"))
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(workspacePath, "gitmodules"), filepath.Join(workspacePath, ".gitmodules"))
	if err == nil {
		subModules := struct {
			SubModule map[string]*struct {
				Path string
				Url string
			}
		}{}
		err := gcfg.ReadFileInto(&subModules, filepath.Join(workspacePath, ".gitmodules"))
		if err != nil {
			return err
		}
		for _, subModule := range subModules.SubModule {
			err = os.Rename(filepath.Join(workspacePath, subModule.Path, ".gitted"),
			                filepath.Join(workspacePath, subModule.Path, ".git"))
		}
	}
	return nil
}

func CleanupWorkspace() error {
	return CleanupFixture()
}

func PrepareEmptyWorkDir(workspacePath string) error {
	if backupWorkspacePath != "" {
		return errors.New("Workspace is initialized")
	}
	_, err := os.Stat(workspacePath)
	if !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Workspace name %s exists already", workspacePath))
	}
	currentWorkspacePath = workspacePath
	return os.MkdirAll(workspacePath, 0777)
}

func CleanupEmptyWorkDir() error {
	if currentWorkspacePath == "" {
		return errors.New("Workspace is not initialized")
	}
	os.RemoveAll(currentWorkspacePath)
	currentWorkspacePath = ""
	return nil
}