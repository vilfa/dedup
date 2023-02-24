package cli

import (
	"sync"
	"time"

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
		c.Log.Print("expected file path")
		return 1
	}

	mshFt := util.Json
	if len(args) == 2 {
		if args[1] == "yaml" {
			mshFt = util.Yaml
		} else if args[1] == "json" {
			mshFt = util.Json
		} else {
			c.Log.Printf("invalid marshall format")
			return 1
		}
	}

	ds, err := scan.NewDirStat(args[0], mshFt)
	if err != nil {
		c.Log.Printf("dry run error: %v", err)
		return 1
	}

	err = ds.Read()
	if err != nil {
		c.Log.Printf("dry run error: %v", err)
		c.Log.Printf("last run was never")
	} else {
		c.Log.Printf("last run was at %v", ds.Ts())
	}

	err = ds.Stat()
	if err != nil {
		c.Log.Printf("dry run error: %v", err)
		return 1
	}

	err = ds.Write()
	if err != nil {
		c.Log.Printf("dry run error: %v", err)
		return 1
	}

	c.Log.Print("created dirstat file")

	h := scan.NewFileHasher(ds)

	chF, chE, err := h.Run()
	if err != nil {
		c.Log.Printf("file hasher error: %v", err)
		return 1
	}
	chD := h.DoneNotify()

	var fhs []scan.FileStatImpl

	prog := struct {
		wg   sync.WaitGroup
		m    sync.Mutex
		n    int
		nall int
		done bool
	}{wg: sync.WaitGroup{}, m: sync.Mutex{}, n: 0, nall: ds.LenFiles(), done: false}

	prog.wg.Add(2)

	go func() {
		for {
			select {
			case f := <-chF:
				fhs = append(fhs, f)
				prog.m.Lock()
				prog.n = len(fhs)
				prog.m.Unlock()
			case err := <-chE:
				c.Log.Printf("file hash processing error: %v", err)
			case done := <-chD:
				if done {
					c.Log.Print("file hash operation done")
					prog.m.Lock()
					prog.done = true
					prog.m.Unlock()
					prog.wg.Done()
					return
				}
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			prog.m.Lock()
			c.Log.Printf("processed %v/%v files", prog.n, prog.nall)
			if prog.done {
				prog.m.Unlock()
				prog.wg.Done()
				return
			}
			prog.m.Unlock()
		}
	}()

	prog.wg.Wait()

	// for _, fh := range fhs {
	// fh.Write(c.Log.Writer())
	// }

	return 0
}
