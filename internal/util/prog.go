package util

import (
	"sync"
	"time"
)

type Progressor struct {
	wg    sync.WaitGroup
	m     sync.Mutex
	t     time.Time
	uProg int
	uDone int
	done  bool
}

func NewProgressor(unitDone int) *Progressor {
	return &Progressor{
		wg:    sync.WaitGroup{},
		m:     sync.Mutex{},
		t:     time.Now(),
		uProg: 0,
		uDone: unitDone,
		done:  false,
	}
}

func (p *Progressor) Inc() bool {
	p.m.Lock()
	defer p.m.Unlock()
	if p.uProg >= p.uDone || p.done {
		p.done = true
		return false
	}
	p.uProg++
	return true
}

func (p *Progressor) Done() {
	p.m.Lock()
	defer p.m.Unlock()
	p.done = true
}

func (p *Progressor) Progress() (int, int, float64, bool) {
	p.m.Lock()
	defer p.m.Unlock()
	return p.uProg, p.uDone, float64(p.uProg) / float64(p.uDone) * 100.0, p.done
}

func (p *Progressor) PercentProgress() float64 {
	p.m.Lock()
	defer p.m.Unlock()
	return float64(p.uProg) / float64(p.uDone) * 100.0
}

func (p *Progressor) UnitProgress() int {
	p.m.Lock()
	defer p.m.Unlock()
	return p.uProg
}

func (p *Progressor) UnitDone() int {
	p.m.Lock()
	defer p.m.Unlock()
	return p.uDone
}

func (p *Progressor) IsDone() bool {
	p.m.Lock()
	defer p.m.Unlock()
	return p.done
}

func (p *Progressor) Elapsed() time.Duration {
	return time.Since(p.t)
}

func (p *Progressor) WgAdd(n int) {
	p.wg.Add(n)
}

func (p *Progressor) WgDone() {
	p.wg.Done()
}

func (p *Progressor) WgWait() {
	p.wg.Wait()
}
