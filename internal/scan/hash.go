package scan

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
)

type MatcherFn func([]byte) (bool, types.Type)

var defaultMatchers matchers.Map = make(matchers.Map)

type fileHasher struct {
	wg        sync.WaitGroup
	wc        int
	dir       *dirStat
	fQueue    []string
	matcherFn MatcherFn
	running   bool
	doneCh    chan bool
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
	Run()
	SetMatcher(MatcherFn)
	SetWorkerCount(int)
	DoneNotify() chan bool
}

func NewFileHasher(dirStat *dirStat, fileQueue []string) FileHasherImpl {
	return &fileHasher{
		wg:        sync.WaitGroup{},
		wc:        100,
		dir:       dirStat,
		fQueue:    fileQueue,
		matcherFn: defaultMatcher,
		running:   false,
		doneCh:    make(chan bool, 1),
	}
}

func NewFileHasherWithMatcher(dirStat *dirStat, fileQueue []string, fn MatcherFn) FileHasherImpl {
	return &fileHasher{
		wg:        sync.WaitGroup{},
		wc:        100,
		dir:       dirStat,
		fQueue:    fileQueue,
		matcherFn: fn,
		running:   false,
		doneCh:    make(chan bool, 1),
	}
}

func (fh *fileHasher) Run() {
	fh.running = true
	fh.doneCh <- false

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
					chErr <- fmt.Errorf("could not read file or weird file: %v, %v", fname, err)
				}

				if ok, typ := fh.matcherFn(buf); ok {
					f.Seek(0, 0)

					hasher := sha256.New()
					if _, err := io.Copy(hasher, f); err != nil {
						chErr <- fmt.Errorf("could not hash file: %v, %v", fname, err)
					}
					fstat, err := NewFileStat(fh.dir, fname, hasher.Sum(nil), typ)
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
		for _, fname := range fh.fQueue {
			chFile <- fname
		}
		close(chFile)
		fh.wg.Wait()
		close(chHash)
		close(chErr)
	}()

	fh.running = false
	fh.doneCh <- true
}

func (fh *fileHasher) SetMatcher(fn MatcherFn) {
	fh.matcherFn = fn
}

func (fh *fileHasher) SetWorkerCount(cnt int) {
	fh.wc = cnt
}

func (fh *fileHasher) DoneNotify() chan bool {
	return fh.doneCh
}
