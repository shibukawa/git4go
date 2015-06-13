package git4go

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var dirListCache map[string][]string = make(map[string][]string)

func findInDirList(name string, label string) (string, error) {
	dirs, ok := dirListCache[label]
	if !ok {
		switch label {
		case "system":
			dirs = guessSystemFile()
		case "global":
			dirs = guessGlobalFile()
		case "global/xdg":
			dirs = guessXDGFile()
		case "template":
			dirs = guessTemplateFile()
		}
		dirListCache[label] = dirs
	}
	for _, dir := range dirs {
		path := dir
		if name != "" {
			path = filepath.Join(path, name)
		}
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", errors.New(fmt.Sprintf("The %s file '%s' doesn't exist", label, name))
}
