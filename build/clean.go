package main

import (
	"os"

	"github.com/goyek/goyek/v2"
)

var _ = goyek.Define(goyek.Task{
	Name:  "clean",
	Usage: "Remove generated files",
	Action: func(a *goyek.A) {
		if _, err := os.Stat("bin"); os.IsNotExist(err) {
			return
		}

		a.Log("rm bin/")
		if err := os.RemoveAll("bin"); err != nil {
			a.Error(err)
		}
	},
})
