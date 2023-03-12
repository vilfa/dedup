package cli

import (
	"log"

	"github.com/mitchellh/cli"
)

type OpMode uint

const (
	Index OpMode = iota
	DryRun
	Export
	Resolve
)

type IndexCommand struct {
	*baseCommand
}

type DryRunCommand struct {
	*baseCommand
}

type ExportCommand struct {
	*baseCommand
}

type ResolveCommand struct {
	*baseCommand
}

func NewCommandFactory(logger *log.Logger) func() map[string]cli.CommandFactory {
	base := NewBaseCommand(logger)
	return func() map[string]cli.CommandFactory {
		return map[string]cli.CommandFactory{
			"index": func() (cli.Command, error) {
				base.name = "index"
				return IndexCommand{base}, nil
			},
			"dryrun": func() (cli.Command, error) {
				base.name = "dryrun"
				return DryRunCommand{base}, nil
			},
			"export": func() (cli.Command, error) {
				base.name = "export"
				return ExportCommand{base}, nil
			},
			"resolve": func() (cli.Command, error) {
				base.name = "resolve"
				return ResolveCommand{base}, nil
			},
		}
	}
}
