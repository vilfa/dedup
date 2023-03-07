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

	var err error
	defer func() {
		if err != nil {
			c.Log.Printf("dryrun error: %v", err)
		} else {
			c.Log.Printf("dryrun success")
		}
	}()

	ds, err := scan.NewDirStat(args[0], util.Json)
	if err != nil {
		return 1
	}

	err = ds.Read()
	if err != nil {
		c.Log.Printf("last run was never")
	} else {
		c.Log.Printf("last run was at %v", ds.Timestamp())
	}

	err = ds.Stat()
	if err != nil {
		return 1
	}

	err = ds.Write()
	if err != nil {
		return 1
	}

	h := scan.NewFileHasher(ds)
	chRun, chDone, chErr, err := h.Run()
	if err != nil {
		return 1
	}

	var processed []scan.FileStatImpl
	progress := struct {
		wg   sync.WaitGroup
		m    sync.Mutex
		n    int
		na   int
		done bool
		t    time.Time
	}{wg: sync.WaitGroup{}, m: sync.Mutex{}, n: 0, na: ds.NFiles(), done: false, t: time.Now()}

	progress.wg.Add(2)

	go func() {
		for {
			select {
			case f := <-chRun:
				processed = append(processed, f)
				progress.m.Lock()
				progress.n++
				progress.m.Unlock()
			case err := <-chErr:
				if err != nil {
					c.Log.Printf("processing error: %v", err)
				}
			case done := <-chDone:
				if done {
					progress.m.Lock()
					progress.done = true
					progress.m.Unlock()
					progress.wg.Done()
					return
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			progress.m.Lock()
			c.Log.Printf("processed %v/%v files", progress.n, progress.na)
			if progress.done {
				c.Log.Printf("processing done in %v", time.Since(progress.t))
				progress.m.Unlock()
				progress.wg.Done()
				return
			}
			progress.m.Unlock()
		}
	}()

	progress.wg.Wait()

	// for _, fh := range fhs {
	// fh.Write(c.Log.Writer())
	// }

	return 0
}
