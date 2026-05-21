package util

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Bold(true)
	pctStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Bold(true)
	sizeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	rateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	trackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	spinnerLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	elapsedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5484D")).Bold(true)
	envRedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5484D")).Bold(true)
	gcpLabelSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	gcpValueSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
)

// GCPBanner returns a one-line "GCP project: NAME" banner suitable for
// printing before interactive prompts.
func GCPBanner(project string) string {
	return gcpLabelSty.Render("GCP project:") + " " + gcpValueSty.Render(project)
}

const (
	cursorHide = "\033[?25l"
	cursorShow = "\033[?25h"
)

// EnvRed returns the environment name styled red+bold, used to draw attention
// to acceptance/production targets.
func EnvRed(env string) string {
	return envRedStyle.Render(env)
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func termCols() int {
	w, _, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	// Leave one column free; some terminals auto-wrap on the last column,
	// which corrupts cursor-based redraws.
	return w - 1
}

func gradientStyleAt(i, total int) lipgloss.Style {
	if total <= 1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6366F1"))
	}
	t := float64(i) / float64(total-1)
	r := int(0x81 + t*(0x63-0x81))
	g := int(0x8C + t*(0x66-0x8C))
	b := int(0xF8 + t*(0xF1-0xF8))
	return lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b)))
}

// ProgressWriter wraps an io.Writer and renders a styled progress bar on
// stderr as bytes flow through. Call Finish or FinishFail when done.
//
// If a header is set via SetHeader, the bar renders on two lines: the header
// text on the first line with the percentage right-anchored, and the bar
// itself (with size and rate) on the second.
type ProgressWriter struct {
	dst      io.Writer
	label    string
	header   string
	total    int64
	sent     int64
	start    time.Time
	last     time.Time
	tty      bool
	rendered bool
}

func NewProgressWriter(dst io.Writer, label string, total int64) *ProgressWriter {
	return &ProgressWriter{dst: dst, label: label, total: total, start: time.Now(), tty: isTTY()}
}

func (p *ProgressWriter) SetHeader(header string) *ProgressWriter {
	p.header = header
	return p
}

func (p *ProgressWriter) Write(b []byte) (int, error) {
	n, err := p.dst.Write(b)
	p.sent += int64(n)
	if !p.rendered {
		p.render()
		p.last = time.Now()
	} else if p.tty && time.Since(p.last) >= 50*time.Millisecond {
		p.render()
		p.last = time.Now()
	}
	return n, err
}

func (p *ProgressWriter) Finish() {
	if p.tty {
		p.sent = p.total
		p.render()
		fmt.Fprint(os.Stderr, "\n"+cursorShow)
	} else {
		label := p.header
		if label == "" {
			label = p.label
		}
		if label == "" {
			label = "Done"
		}
		fmt.Fprintf(os.Stderr, "%s: done in %.1fs\n", label, time.Since(p.start).Seconds())
	}
}

func (p *ProgressWriter) FinishFail() {
	if p.tty && p.rendered {
		fmt.Fprint(os.Stderr, "\n"+cursorShow)
	}
}

func (p *ProgressWriter) render() {
	if !p.tty {
		if !p.rendered {
			if p.header != "" {
				fmt.Fprintln(os.Stderr, p.header)
			} else if p.label != "" {
				fmt.Fprintf(os.Stderr, "%s %s...\n", p.label, formatBytes(p.total))
			} else {
				fmt.Fprintf(os.Stderr, "%s...\n", formatBytes(p.total))
			}
			p.rendered = true
		}
		return
	}

	var pct float64
	if p.total > 0 {
		pct = float64(p.sent) / float64(p.total)
	}
	if pct > 1 {
		pct = 1
	}
	elapsed := time.Since(p.start).Seconds()
	var rate float64
	if elapsed > 0 {
		rate = float64(p.sent) / elapsed
	}

	pctStr := pctStyle.Render(fmt.Sprintf("%5.1f%%", pct*100))
	suffix := " " +
		sizeStyle.Render(formatBytesPair(p.sent, p.total)) + " " +
		rateStyle.Render(formatRate(rate))

	if p.header != "" {
		pad := termCols() - lipgloss.Width(p.header) - lipgloss.Width(pctStr)
		if pad < 1 {
			pad = 1
		}
		headerLine := p.header + strings.Repeat(" ", pad) + pctStr

		barWidth := termCols() - lipgloss.Width(suffix)
		if barWidth < 10 {
			barWidth = 10
		}
		barLine := buildBar(pct, barWidth) + suffix

		if p.rendered {
			fmt.Fprintf(os.Stderr, "\033[A\r%s\033[K\n\r%s\033[K", headerLine, barLine)
		} else {
			fmt.Fprintf(os.Stderr, "%s%s\n%s", cursorHide, headerLine, barLine)
			p.rendered = true
		}
		return
	}

	prefix := pctStr + " "
	if p.label != "" {
		prefix = labelStyle.Render(p.label) + " " + prefix
	}
	width := termCols() - lipgloss.Width(prefix) - lipgloss.Width(suffix)
	if width < 10 {
		width = 10
	}
	if !p.rendered {
		fmt.Fprint(os.Stderr, cursorHide)
		p.rendered = true
	}
	fmt.Fprintf(os.Stderr, "\r%s%s%s\033[K", prefix, buildBar(pct, width), suffix)
}

func buildBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	var bar strings.Builder
	for i := 0; i < filled; i++ {
		bar.WriteString(gradientStyleAt(i, width).Render("━"))
	}
	for i := filled; i < width; i++ {
		bar.WriteString(trackStyle.Render("─"))
	}
	return bar.String()
}

func formatBytes(b int64) string {
	div, suffix := bytesUnit(b)
	return fmt.Sprintf("%.1f %s", float64(b)/div, suffix)
}

func formatBytesPair(current, total int64) string {
	div, suffix := bytesUnit(total)
	return fmt.Sprintf("%.1f/%.1f %s", float64(current)/div, float64(total)/div, suffix)
}

func bytesUnit(b int64) (float64, string) {
	switch {
	case b >= 1<<30:
		return float64(int64(1) << 30), "GB"
	case b >= 1<<20:
		return float64(int64(1) << 20), "MB"
	case b >= 1<<10:
		return float64(int64(1) << 10), "KB"
	}
	return 1, "B"
}

func formatRate(bytesPerSec float64) string {
	switch {
	case bytesPerSec >= 1<<30:
		return fmt.Sprintf("%.1f GB/s", bytesPerSec/float64(int64(1)<<30))
	case bytesPerSec >= 1<<20:
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/float64(int64(1)<<20))
	case bytesPerSec >= 1<<10:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/float64(int64(1)<<10))
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

// Spinner renders an animated braille spinner on stderr while an
// indeterminate operation runs. Start with StartSpinner; call Stop or
// StopFail when done.
type Spinner struct {
	label   string
	stop    chan struct{}
	stopped chan struct{}
	start   time.Time
	tty     bool
}

func StartSpinner(label string) *Spinner {
	s := &Spinner{
		label:   label,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		start:   time.Now(),
		tty:     isTTY(),
	}
	if s.tty {
		fmt.Fprint(os.Stderr, cursorHide)
		go s.run()
	} else {
		fmt.Fprintf(os.Stderr, "%s...\n", s.label)
		close(s.stopped)
	}
	return s
}

func (s *Spinner) run() {
	defer close(s.stopped)
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	i := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			renderSpinnerLine(
				spinnerStyle.Render(string(frames[i%len(frames)])),
				s.label,
				time.Since(s.start),
			)
			i++
		}
	}
}

func renderSpinnerLine(symbol, label string, elapsed time.Duration) {
	left := symbol + " " + spinnerLabel.Render(label)
	right := elapsedStyle.Render(fmt.Sprintf("(%.1fs)", elapsed.Seconds()))
	pad := termCols() - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	fmt.Fprintf(os.Stderr, "\r%s%s%s\033[K", left, strings.Repeat(" ", pad), right)
}

func (s *Spinner) Stop()     { s.finish("✓", doneStyle) }
func (s *Spinner) StopFail() { s.finish("✗", failStyle) }

func (s *Spinner) finish(symbol string, style lipgloss.Style) {
	if s.tty {
		close(s.stop)
		<-s.stopped
		renderSpinnerLine(style.Render(symbol), s.label, time.Since(s.start))
		fmt.Fprint(os.Stderr, "\n"+cursorShow)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s (%.1fs)\n", symbol, s.label, time.Since(s.start).Seconds())
	}
}
