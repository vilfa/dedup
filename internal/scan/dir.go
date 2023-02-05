package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/vilfa/dedup/internal/util"
)

const (
	dirStatFname = ".dedup_ds"
)

type dirStat struct {
	Path         string
	Size         int
	FileCount    int
	DirCount     int
	SpecCount    int
	Timestamp    int64
	marshallType util.MarshallType
}

type DirStatImpl interface {
	Read() error
	Write() error
	Stat() error
	Ts() time.Time
}

func NewDirStat(dir string, mTyp util.MarshallType) (DirStatImpl, error) {
	return &dirStat{
		Path:         dir,
		Size:         0,
		FileCount:    0,
		DirCount:     0,
		Timestamp:    0,
		marshallType: mTyp,
	}, nil
}

func (d *dirStat) Read() error {
	fpath := path.Join(d.Path, dirStatFname)
	if !fs.ValidPath(fpath) {
		return fmt.Errorf("invalid dir info path: %v", fpath)
	}

	buf, err := os.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("could not read dir info: %v", err)
	}

	err = util.Unmarshall(d.marshallType, buf, d)
	if err != nil {
		if d.marshallType == util.Json {
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
	fpath := path.Join(d.Path, dirStatFname)
	if !fs.ValidPath(fpath) {
		return fmt.Errorf("invalid dir info path: %v", fpath)
	}

	buf, err := util.Marshall(d.marshallType, d)
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
	if !fs.ValidPath(d.Path) {
		return errors.New("invalid wdir")
	}

	dir, err := os.ReadDir(d.Path)
	if err != nil {
		return errors.New("could not open wdir")
	}

	d.Size = len(dir)
	d.DirCount = 0
	d.FileCount = 0
	d.SpecCount = 0
	d.Timestamp = time.Now().UnixMilli()
	for _, dent := range dir {
		if dent.Type().IsRegular() {
			d.FileCount++
		} else if dent.Type().IsDir() {
			d.DirCount++
		} else {
			d.SpecCount++
		}
	}

	return nil
}

func (d *dirStat) Ts() time.Time {
	return time.UnixMilli(d.Timestamp)
}
