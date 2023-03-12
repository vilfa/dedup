package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/vilfa/dedup/internal/scan"
	"github.com/vilfa/dedup/internal/util"
)

func (c IndexCommand) Help() string {
	return ""
}

func (c IndexCommand) Synopsis() string {
	return "Build an index of image duplicates."
}

func (c IndexCommand) Run(args []string) int {
	var err error
	defer DeferredExit(c.baseCommand, &err)

	if len(args) < 1 {
		err = fmt.Errorf("expected file path")
		return 1
	}

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
	chData, chErr, chDone, err := h.Run()
	if err != nil {
		return 1
	}

	var out []scan.FileStatImpl
	progress := util.NewProgressor(ds.NFiles())
	progress.WgAdd(2)

	go func() {
		defer progress.WgDone()

		for {
			select {
			case data := <-chData:
				out = append(out, data)
				progress.Inc()
			case err := <-chErr:
				if err != nil {
					c.Log.Printf("processing error: %v", err)
				}
			case <-chDone:
				progress.Done()
				return
			}
		}
	}()

	go func() {
		defer progress.WgDone()

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			pCurr, pAll, pPerc, pDone := progress.Progress()

			c.Log.Printf("processed %v/%v (%.2f%%) files", pCurr, pAll, pPerc)
			if pDone {
				c.Log.Printf("processing done in %v", progress.Elapsed())
				return
			}
		}
	}()

	progress.WgWait()

	dups := make(map[string][]string)
	for _, f := range out { // TODO: Panic here sometimes, goroutine + memory related probably.
		dups[f.Sha256()] = append(dups[f.Sha256()], f.Path())
	}

	bDups, err := util.Marshall(util.Json, dups)
	if err != nil {
		return 1
	}

	err = os.WriteFile("temp.json", bDups, 0644)
	if err != nil {
		return 1
	}

	return 0
}
