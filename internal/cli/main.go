package cli

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/vilfa/dedup/internal/log"
)

func Main(args []string) int {
	logger := log.Default()
	commands := NewCommandFactory(logger)()
	cli := &cli.CLI{
		Name:        AppName,
		Version:     AppVersion,
		Args:        args[1:],
		HelpWriter:  os.Stdout,
		ErrorWriter: os.Stderr,
		Commands:    commands,
	}

	exitCode, err := cli.Run()
	if err != nil {
		logger.Printf("failed to run command: %s\n", err)
		return exitCode
	}

	return exitCode
}
