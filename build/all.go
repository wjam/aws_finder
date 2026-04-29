package main

import "github.com/goyek/goyek/v3"

var pipelineAll = goyek.Define(goyek.Task{
	Name:  "all",
	Usage: "build pipeline",
	Deps: goyek.Deps{
		stageInit,
		stageTest,
		stageBuild,
	},
})
