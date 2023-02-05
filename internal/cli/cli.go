package cli

import (
	"context"
	"log"
)

const (
	AppName    string = "dedup"
	AppVersion string = "0.1.0"
)

type baseCommand struct {
	Ctx context.Context

	Log *log.Logger
}

func NewBaseCommand(logger *log.Logger) *baseCommand {
	return &baseCommand{
		Ctx: context.Background(),
		Log: logger,
	}
}
