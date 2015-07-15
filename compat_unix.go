// +build dragonfly freebsd linux nacl netbsd openbsd solaris

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

var defaultBoolConfig map[string]bool = map[string]bool{
	"core.symlinks":          true,
	"core.ignorecase":        false,
	"core.filemode":          true,
	"core.ignorestat":        false,
	"core.trustctime":        true,
	"core.abbrev":            true,
	"core.precomposeunicode": true,
	"core.logallrefupdates":  true,
	"core.protectHFS":        false,
	"core.protectNTFS":       false,
}

var defaultStringConfig map[string]string = map[string]string{
	"core.autocrlf": "false",
	"core.eol":      "crlf",
}
