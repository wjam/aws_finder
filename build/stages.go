package main

import "github.com/goyek/goyek/v3"

var stageInit = goyek.Define(goyek.Task{
	Name:  "init",
	Usage: "init stage",
	Deps: goyek.Deps{
		goGenerate,
	},
})

var stageTest = goyek.Define(goyek.Task{
	Name:  "test",
	Usage: "test stage",
	Deps: goyek.Deps{
		goTest,
		golangciLint,
		goModTidyDiff,
	},
})

var stageBuild = goyek.Define(goyek.Task{
	Name:  "build",
	Usage: "build stage",
	Deps: goyek.Deps{
		goBuild,
	},
})
