package main

import (
	"os"
	"path/filepath"

	"github.com/vilfa/dedup/internal/cli"
)

func main() {
	os.Args[0] = filepath.Base(os.Args[0])
	os.Exit(cli.Main(os.Args))
}
