package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/shibukawa/git4go"
	"os"
)

func CmdLsTree(c *cli.Context) {
	repo, err := git4go.OpenRepositoryExtended(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(c.Args()) == 0 {
		cli.ShowSubcommandHelp(c)
	} else {
		var tree *git4go.Tree
		var commit *git4go.Commit

		oid, err := git4go.NewOid(c.Args().First())

		if err == nil {
			obj, err := repo.Lookup(oid)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if obj.Type() == git4go.ObjectTree {
				tree = obj.(*git4go.Tree)
			} else if obj.Type() == git4go.ObjectCommit {
				commit = obj.(*git4go.Commit)
			} else {
				os.Stderr.WriteString("fatal: not a tree object")
				os.Exit(1)
			}
		} else {
			ref, err := repo.DwimReference(c.Args().First())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			resolved, err := ref.Resolve()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			commit, err = repo.LookupCommit(resolved.Target())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		if commit != nil {
			tree, err = commit.Tree()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		for _, entry := range tree.Entries {
			fileMode := fmt.Sprintf("%06o", int(entry.Filemode))
			fmt.Printf("%s %s %s\t%s\n", fileMode, entry.Type.String(), entry.Id.String(), entry.Name)
		}
	}
}
