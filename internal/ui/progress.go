package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Progress renders a single-line progress indicator to w (typically stderr).
// It is a no-op when w is not a terminal, so piped output stays clean.
type Progress struct {
	mu      sync.Mutex
	w       io.Writer
	enabled bool
	total   int
	done    int
	label   string
}

// NewProgress returns a Progress writing to w. Progress is only shown when w
// is a character device (a terminal).
func NewProgress(w io.Writer, total int, label string) *Progress {
	enabled := false
	if f, ok := w.(*os.File); ok {
		if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
			enabled = true
		}
	}
	return &Progress{w: w, enabled: enabled, total: total, label: label}
}

// Inc advances the counter by one and redraws the line.
func (p *Progress) Inc() {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	p.done++
	p.render()
	p.mu.Unlock()
}

func (p *Progress) render() {
	fmt.Fprintf(p.w, "\r\x1b[2K[%d/%d] %s", p.done, p.total, p.label)
}

// Done clears the progress line.
func (p *Progress) Done() {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	fmt.Fprint(p.w, "\r\x1b[2K")
	p.mu.Unlock()
}
