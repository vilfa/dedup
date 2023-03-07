package scan

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
)

type MatcherFn func([]byte) (bool, types.Type)

var (
	defaultMatchers matchers.Map = make(matchers.Map)
	goroutineCount  int          = 250
)

type fileHasher struct {
	wg        sync.WaitGroup
	wc        int
	dir       DirStatImpl
	matcherFn MatcherFn
	running   bool
	chDone    chan bool
}

func init() {
	defaultMatchers = matchers.Image
	for k, v := range matchers.Video {
		defaultMatchers[k] = v
	}
}

func defaultMatcher(buf []byte) (bool, types.Type) {
	typ := filetype.MatchMap(buf, defaultMatchers)
	return typ != types.Unknown, typ
}

type FileHasherImpl interface {
	Run() (chan FileStatImpl, chan bool, chan error, error)
	SetMatcher(MatcherFn)
	SetWorkerCount(int)
}

func NewFileHasher(dirStat DirStatImpl) FileHasherImpl {
	return &fileHasher{
		wg:        sync.WaitGroup{},
		wc:        goroutineCount,
		dir:       dirStat,
		matcherFn: defaultMatcher,
		running:   false,
		chDone:    make(chan bool, 1),
	}
}

func NewFileHasherWithMatcher(dirStat DirStatImpl, fn MatcherFn) FileHasherImpl {
	return &fileHasher{
		wg:        sync.WaitGroup{},
		wc:        goroutineCount,
		dir:       dirStat,
		matcherFn: fn,
		running:   false,
		chDone:    make(chan bool, 1),
	}
}

func (fh *fileHasher) Run() (chan FileStatImpl, chan bool, chan error, error) {
	fh.running = true
	fh.chDone <- false

	fQueue, err := fh.dir.Files()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not read dir: %v", err)
	}

	chFile := make(chan string)
	chHash := make(chan FileStatImpl)
	chErr := make(chan error)

	for i := 0; i < fh.wc; i++ {
		fh.wg.Add(1)
		go func() {
			defer fh.wg.Done()
			for fname := range chFile {
				f, err := os.Open(fname)
				if err != nil {
					chErr <- fmt.Errorf("could not open file: %v, %v", fname, err)
					return
				}

				buf := make([]byte, 261)
				if n, err := f.Read(buf); n != 261 || err != nil {
					chErr <- fmt.Errorf("skipping abnormal file: %v, is_tiny: %v, err: %v", fname, n < 261, err)
					f.Close()
					return
				}

				if ok, typ := fh.matcherFn(buf); ok {
					f.Seek(0, 0)

					hasher := sha256.New()
					if _, err := io.Copy(hasher, f); err != nil {
						chErr <- fmt.Errorf("could not hash file: %v, %v", fname, err)
					}
					fstat, err := NewFileStat(fh.dir, path.Base(fname), hasher.Sum(nil), typ)
					if err != nil {
						chErr <- fmt.Errorf("could not hash file: %v", err)
					}
					chHash <- fstat
				}

				f.Close()
			}
		}()
	}

	go func() {
		for _, f := range fQueue {
			chFile <- filepath.Join(fh.dir.Path(), f.Name())
		}
		close(chFile)
		fh.wg.Wait()
		close(chHash)
		close(chErr)
		fh.running = false
		fh.chDone <- true
	}()

	return chHash, fh.chDone, chErr, nil
}

func (fh *fileHasher) SetMatcher(fn MatcherFn) {
	fh.matcherFn = fn
}

func (fh *fileHasher) SetWorkerCount(count int) {
	fh.wc = count
}

func (fh *fileHasher) DoneNotify() chan bool {
	return fh.chDone
}
