package pipeline

import (
	"sync"
	"time"

	"pdf_watermark_remover/internal/logutil"
)

type Progress struct {
	total int
	done  int
	start time.Time
	mu    sync.Mutex
}

func NewProgress(total int) *Progress {
	return &Progress{total: total, start: time.Now()}
}

func (p *Progress) Tick() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	logutil.Printf("processed %d/%d\n", p.done, p.total)
	return p.done
}

func (p *Progress) Duration() time.Duration {
	return time.Since(p.start)
}
