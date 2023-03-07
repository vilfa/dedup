package cli

import (
	"os"
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
	chData, chDone, chErr, err := h.Run()
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
				c.Log.Printf("processing error: %v", err)
				progress.Inc()
			case <-chDone:
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

			c.Log.Printf("processed %v/%v (%v%%) files", pCurr, pAll, pPerc)
			if pDone {
				c.Log.Printf("processing done in %v", progress.Elapsed())
				return
			}
		}
	}()

	progress.WgWait()

	dups := make(map[string][]string)
	for _, f := range out { // TODO: Panic here.
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
