package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/shibukawa/git4go"
	"os"
)

func CmdCatFile(c *cli.Context) {
	repo, err := git4go.OpenRepositoryExtended(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(c.Args()) < 2 {
		cli.ShowSubcommandHelp(c)
	} else {
		objType := git4go.TypeString2Type(c.Args().First())
		if objType == git4go.ObjectBad {
			fmt.Fprintln(os.Stderr, `fatal: invalid object type "bad"`)
			os.Exit(1)
		}
		odb, err := repo.Odb()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		oid, err := git4go.NewOid(c.Args()[1])
		if err != nil {
			ref, err := repo.DwimReference(c.Args()[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			resolved, err := ref.Resolve()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			oid = resolved.Target()
		}
		obj, err := odb.Read(oid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if obj.Type == git4go.ObjectCommit && objType == git4go.ObjectTree {
			commit, _ := repo.LookupCommit(oid)
			obj, err = odb.Read(commit.TreeId())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		if obj.Type != objType {
			fmt.Fprintf(os.Stderr, "fatal: got cat-file: %s: bad file\n", c.Args()[1])
			os.Exit(1)
		}
		os.Stdout.Write(obj.Data)
	}
}

func CompletionCatFile(c *cli.Context) {
	fmt.Println("bash completeion cat-file")
	types := []string{
		"blob",
		"commit",
		"tag",
		"tree",
	}
	if len(c.Args()) != 0 {
		for _, typeStr := range types {
			fmt.Println(typeStr)
		}
	}
}
