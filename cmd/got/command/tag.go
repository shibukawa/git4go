package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/shibukawa/git4go"
	"os"
)

func CmdTag(c *cli.Context) {
	repo, err := git4go.OpenRepositoryExtended(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch len(c.Args()) {
	case 0:
		tags, err := repo.ListTag()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
	}
}
