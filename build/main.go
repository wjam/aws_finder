package main

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/boot"
)

func init() {
	goyek.SetDefault(pipelineAll)
}

func main() {
	boot.Main()
}
