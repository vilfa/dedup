package scan

import (
	"fmt"
	"io"
	"path"

	"github.com/h2non/filetype/types"
	"github.com/vilfa/dedup/internal/util"
)

type fileStat struct {
	Parent DirStatImpl
	Path   string
	Type   types.Type
	Hash   []byte
}

type FileStatImpl interface {
	Write(w io.Writer) error
}

func NewFileStat(parent DirStatImpl, name string, hash []byte, typ types.Type) (FileStatImpl, error) {
	fpath := path.Join(parent.Path(), name)

	return &fileStat{
		Parent: parent,
		Path:   fpath,
		Hash:   hash,
		Type:   typ,
	}, nil
}

func (f *fileStat) Write(w io.Writer) error {
	buf, err := util.Marshall(f.Parent.MarshalType(), f)
	if err != nil {
		return fmt.Errorf("could not marshall: %v", err)
	}

	n, err := w.Write(buf)
	if n != len(buf) || err != nil {
		return fmt.Errorf("could not write to buffer: %v", err)
	}

	return nil
}
