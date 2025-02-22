package main

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var golangciLint = goyek.Define(goyek.Task{
	Name:  "golangci-lint",
	Usage: "golangci-lint",
	Action: func(a *goyek.A) {
		cmd.Exec(a, "go tool -modfile=./tools/golangci-lint/go.mod golangci-lint run")
	},
})

var _ = goyek.Define(goyek.Task{
	Name:  "golangci-lint-fix",
	Usage: "golangci-lint-fix",
	Action: func(a *goyek.A) {
		cmd.Exec(a, "go tool -modfile=./tools/golangci-lint/go.mod golangci-lint run --fix")
	},
})
