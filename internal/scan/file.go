package scan

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/vilfa/dedup/internal/util"
)

type FileType uint

const (
	Jpg FileType = iota
	Png
	Mov
	Mp4
)

type fileStat struct {
	Parent *dirStat
	Path   string
	Type   FileType
	Hash   string
}

type FileStatImpl interface {
	Read() error
	Write(w io.Writer) error
}

func newFileStat(parent *dirStat, name, hash string, typ FileType) (FileStatImpl, error) {
	fpath := path.Join(parent.Path, name)
	if !fs.ValidPath(fpath) {
		return nil, fmt.Errorf("invalid path: %v", fpath)
	}

	return &fileStat{
		Parent: parent,
		Path:   fpath,
		Hash:   hash,
		Type:   typ,
	}, nil
}

func (f *fileStat) Read() error {
	buf, err := os.ReadFile(f.Path)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}

	return util.Unmarshall(f.Parent.marshallType, buf, f)
}

func (f *fileStat) Write(w io.Writer) error {
	buf, err := util.Marshall(f.Parent.marshallType, f)
	if err != nil {
		return fmt.Errorf("could not marshall file: %v", err)
	}

	n, err := w.Write(buf)
	if n != len(buf) || err != nil {
		return fmt.Errorf("error writing to buffer: %v", err)
	}

	return nil
}
