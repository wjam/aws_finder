package main

import (
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
)

var mkdirBin = goyek.Define(goyek.Task{
	Name:  "mkdir-bin",
	Usage: "mkdir bin",
	Action: func(a *goyek.A) {
		s, err := os.Stat("bin")
		if err != nil {
			if !os.IsNotExist(err) {
				a.Error(err)
				return
			}
		} else if !s.IsDir() {
			a.Error(fmt.Errorf("%s is not a directory", s))
			return
		}

		a.Log("mkdir bin/")
		if err := os.MkdirAll("bin", 0750); err != nil {
			a.Error(err)
			return
		}
	},
})
