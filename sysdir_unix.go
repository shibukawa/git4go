// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package git4go

import (
	"os"
	"path/filepath"
)

func guessSystemFile() []string {
	return []string{"/etc"}
}

func guessGlobalFile() []string {
	return []string{os.Getenv("HOME")}
}

func guessXDGFile() []string {
	env := os.Getenv("XDG_CONFIG_HOME")
	if env != "" {
		return []string{filepath.Join(env, "git")}
	} else {
		home := os.Getenv("HOME")
		if home != "" {
			return []string{filepath.Join(home, ".config/git")}
		} else {
			return []string{}
		}
	}
}

func guessTemplateFile() []string {
	return []string{"/usr/share/git-core/templates"}
}
