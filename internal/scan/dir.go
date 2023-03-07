package scan

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"time"

	"github.com/vilfa/dedup/internal/util"
)

const (
	dirStatFname = ".dedup_ds"
)

var (
	cwd string
)

type dirStat struct {
	AbsPath     string
	Size        int
	NumFiles    int
	NumDirs     int
	NumSpecial  int
	Ts          int64
	marshalType util.MarshallType
	dEnts       []fs.DirEntry
}

type DirStatImpl interface {
	Path() string
	Read() error
	Write() error
	Stat() error
	Timestamp() time.Time
	Files() ([]fs.DirEntry, error)
	NFiles() int
	MarshalType() util.MarshallType
}

func init() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Panicf("could not get cwd: %v", err)
	}
}

func resolvePath(p string) string {
	if path.IsAbs(p) {
		return p
	} else {
		return path.Join(cwd, p)
	}
}

func NewDirStat(dir string, mTyp util.MarshallType) (DirStatImpl, error) {
	return &dirStat{
		AbsPath:     resolvePath(dir),
		Size:        0,
		NumFiles:    0,
		NumDirs:     0,
		Ts:          0,
		marshalType: mTyp,
		dEnts:       nil,
	}, nil
}

func (d *dirStat) Path() string {
	return d.AbsPath
}

func (d *dirStat) Read() error {
	fpath := path.Join(d.AbsPath, dirStatFname)

	buf, err := os.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("could not read dir info: %v", err)
	}

	err = util.Unmarshall(d.marshalType, buf, d)
	if err != nil {
		if d.marshalType == util.Json {
			err = util.Unmarshall(util.Yaml, buf, d)
		} else {
			err = util.Unmarshall(util.Json, buf, d)
		}
		if err != nil {
			return fmt.Errorf("could not parse dir info: %v", err)
		}
	}

	return nil
}

func (d *dirStat) Write() error {
	fpath := path.Join(d.AbsPath, dirStatFname)

	buf, err := util.Marshall(d.marshalType, d)
	if err != nil {
		return fmt.Errorf("could not marshall dir info: %v", err)
	}

	err = os.WriteFile(fpath, buf, 0644)
	if err != nil {
		return fmt.Errorf("could not write dir info: %v", err)
	}

	return nil
}

func (d *dirStat) Stat() error {
	dir, err := os.ReadDir(d.AbsPath)
	if err != nil {
		return fmt.Errorf("could not open dir: %v", d.AbsPath)
	}

	d.Size = len(dir)
	d.NumDirs = 0
	d.NumFiles = 0
	d.NumSpecial = 0
	d.Ts = time.Now().UnixMilli()
	for _, dent := range dir {
		if dent.Type().IsRegular() {
			d.NumFiles++
		} else if dent.Type().IsDir() {
			d.NumDirs++
		} else {
			d.NumSpecial++
		}
	}

	return nil
}

func (d *dirStat) Timestamp() time.Time {
	return time.UnixMilli(d.Ts)
}

func (d *dirStat) Files() ([]fs.DirEntry, error) {
	if d.dEnts != nil {
		return d.dEnts, nil
	}

	dir, err := os.ReadDir(d.AbsPath)
	if err != nil {
		return nil, fmt.Errorf("could not open dir: %v", err)
	}

	for _, dent := range dir {
		if dent.Type().IsRegular() {
			d.dEnts = append(d.dEnts, dent)
		}
	}

	return d.dEnts, nil
}

func (d *dirStat) NFiles() int {
	return d.NumFiles
}

func (d *dirStat) MarshalType() util.MarshallType {
	return d.marshalType
}
