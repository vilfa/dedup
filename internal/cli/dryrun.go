package cli

import (
	"github.com/vilfa/dedup/internal/scan"
	"github.com/vilfa/dedup/internal/util"
)

func (c DryRunCommand) Help() string {
	return ""
}

func (c DryRunCommand) Synopsis() string {
	return "Perform a dryrun resolve operation, log everything."
}

func (c DryRunCommand) Run(args []string) int {
	if len(args) < 1 {
		c.base.Log.Print("expected file path")
		return 1
	}

	mshFt := util.Json
	if len(args) == 2 {
		if args[1] == "yaml" {
			mshFt = util.Yaml
		} else if args[1] == "json" {
			mshFt = util.Json
		} else {
			c.base.Log.Printf("invalid marshall format")
			return 1
		}
	}

	ds, err := scan.NewDirStat(args[0], mshFt)
	if err != nil {
		c.base.Log.Printf("dry run error: %v", err)
		return 1
	}

	err = ds.Read()
	if err != nil {
		c.base.Log.Printf("dry run error: %v", err)
		c.base.Log.Printf("last run was never")
	} else {
		c.base.Log.Printf("last run was at %v", ds.Ts())
	}

	err = ds.Stat()
	if err != nil {
		c.base.Log.Printf("dry run error: %v", err)
		return 1
	}

	err = ds.Write()
	if err != nil {
		c.base.Log.Printf("dry run error: %v", err)
		return 1
	}

	c.base.Log.Print("created dirstat file")

	return 0
}
