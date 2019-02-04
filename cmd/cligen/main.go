package main

import (
	"os"

	"github.com/3d0c/cli/pkg"

	_ "github.com/3d0c/cli/cmd/cligen/command"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
