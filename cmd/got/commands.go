package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/shibukawa/got/command"
	"os"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	/*{
		Name:        "add",
		Usage:       "",
		Action:      command.CmdAdd,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "bisect",
		Usage:       "",
		Action:      command.CmdBisect,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "blame",
		Usage:       "",
		Action:      command.CmdBlame,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "branch",
		Usage:       "",
		Action:      command.CmdBranch,
		Flags:       []cli.Flag{},
	},*/
	{
		Name:  "cat-file",
		Usage: "Provide content or type and size information for repository objects",
		Description: `In its first form, the command provides the content or the type of an object in the repository. The type is required unless -t or -p is used to find the object type, or -s is used to find the
   object size, or --textconv is used (which implies type "blob").

   In the second form, a list of objects (separated by linefeeds) is provided on stdin, and the SHA-1, type, and size of each object is printed on stdout.`,
		Action: command.CmdCatFile,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "t",
				Usage: "Instead of the content, show the object type identified by <object>.",
			},
			cli.BoolFlag{
				Name:  "s",
				Usage: "Instead of the content, show the object size identified by <object>.",
			},
			cli.BoolFlag{
				Name:  "e",
				Usage: "Suppress all output; instead exit with zero status if <object> exists and is a valid object.",
			},
			cli.BoolFlag{
				Name:  "p",
				Usage: "Pretty-print the contents of <object> based on its type.",
			},
		},
		BashComplete: command.CompletionCatFile,
	},
	/*{
		Name:        "checkout",
		Usage:       "",
		Action:      command.CmdCheckout,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "clone",
		Usage:       "",
		Action:      command.CmdClone,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "commit",
		Usage:       "",
		Action:      command.CmdCommit,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "diff",
		Usage:       "",
		Action:      command.CmdDiff,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "fetch",
		Usage:       "",
		Action:      command.CmdFetch,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "grep",
		Usage:       "",
		Action:      command.CmdGrep,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "init",
		Usage:       "",
		Action:      command.CmdInit,
		Flags:       []cli.Flag{},
	},

	{
		Name:        "log",
		Usage:       "",
		Action:      command.CmdLog,
		Flags:       []cli.Flag{},
	},

	{
		Name:        "merge",
		Usage:       "",
		Action:      command.CmdMerge,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "mv",
		Usage:       "",
		Action:      command.CmdMv,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "pull",
		Usage:       "",
		Action:      command.CmdPull,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "push",
		Usage:       "",
		Action:      command.CmdPush,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "rebase",
		Usage:       "",
		Action:      command.CmdRebase,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "reset",
		Usage:       "",
		Action:      command.CmdReset,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "rm",
		Usage:       "",
		Action:      command.CmdRm,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "show",
		Usage:       "",
		Action:      command.CmdShow,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "status",
		Usage:       "",
		Action:      command.CmdStatus,
		Flags:       []cli.Flag{},
	},*/
	{
		Name:   "tag",
		Usage:  "Create, list, delete or verify a tag object signed with GPG",
		Action: command.CmdTag,
		Flags:  []cli.Flag{},
	},
	/*{
		Name:        "config",
		Usage:       "",
		Action:      command.CmdConfig,
		Flags:       []cli.Flag{},
	},

	{
		Name:        "fetch",
		Usage:       "",
		Action:      command.CmdFetch,
		Flags:       []cli.Flag{},
	},

	{
		Name:        "submodule",
		Usage:       "",
		Action:      command.CmdSubmodule,
		Flags:       []cli.Flag{},
	},

	{
		Name:        "stash",
		Usage:       "",
		Action:      command.CmdStash,
		Flags:       []cli.Flag{},
	},
	{
		Name:        "remote",
		Usage:       "",
		Action:      command.CmdRemote,
		Flags:       []cli.Flag{},
	},*/
	{
		Name:   "ls-tree",
		Usage:  "List the contents of a tree object",
		Action: command.CmdLsTree,
		Flags:  []cli.Flag{},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
