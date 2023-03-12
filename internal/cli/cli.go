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
	Ctx  context.Context
	Log  *log.Logger
	name string
}

func NewBaseCommand(logger *log.Logger) *baseCommand {
	return &baseCommand{
		Ctx:  context.Background(),
		Log:  logger,
		name: "",
	}
}

func DeferredExit(c *baseCommand, err *error) {
	if *err != nil {
		c.Log.Printf("%v command error: %v", c.name, *err)
	} else {
		c.Log.Printf("%v command success", c.name)
	}
}
