package main

import (
    "fmt"
    "github.com/libgit2/git2go"
)

func main() {
    //a, _ := git.Discover("./testrepo.git", false, []string{})
    //fmt.Println(a, b)
    r, _ := git.OpenRepository("./testrepo")
    oid, _ := git.NewOid("1810dff58d8a660512d4832e740f692884338ccd")
    t, _ := r.LookupTree(oid)
    t.Walk(func(root string, entry *git.TreeEntry) int {
        fmt.Println(root, entry.Name)
        return 0
    })
}
