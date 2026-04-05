package ui

import (
	"fmt"
	"strings"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"
	FgGray    = "\033[90m"

	BgRed   = "\033[41m"
	BgGreen = "\033[42m"
)

func Styled(text string, codes ...string) string {
	prefix := ""
	for _, c := range codes {
		prefix += c
	}
	return prefix + text + Reset
}

func Repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func PadRight(s string, width int) string {
	visible := VisibleLen(s)
	if visible >= width {
		return s
	}
	return s + Repeat(" ", width-visible)
}

func Center(s string, width int) string {
	visible := VisibleLen(s)
	if visible >= width {
		return s
	}
	left := (width - visible) / 2
	right := width - visible - left
	return Repeat(" ", left) + s + Repeat(" ", right)
}

func VisibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}

func ProgressBar(fraction float64, width int) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(fraction * float64(width))

	color := CurrentTheme.Success
	if fraction < 0.3 {
		color = CurrentTheme.Error
	} else if fraction < 0.6 {
		color = CurrentTheme.Warning
	}

	bar := Styled(Repeat("━", filled), color, Bold) + Styled(Repeat("━", width-filled), CurrentTheme.TextDim, Dim)
	return bar
}

func Separator(width int) string {
	return Styled(Repeat("─", width), CurrentTheme.TextDim, Dim)
}

func SparkWPM(wpm float64) string {
	color := CurrentTheme.TextDim
	if wpm >= 80 {
		color = CurrentTheme.Success
	} else if wpm >= 50 {
		color = CurrentTheme.Info
	} else if wpm >= 30 {
		color = CurrentTheme.Warning
	} else if wpm > 0 {
		color = CurrentTheme.Error
	}
	return Styled(fmt.Sprintf("%.0f", wpm), color, Bold)
}

func SparkAccuracy(acc float64) string {
	color := CurrentTheme.Error
	if acc >= 98 {
		color = CurrentTheme.Success
	} else if acc >= 90 {
		color = CurrentTheme.Info
	} else if acc >= 75 {
		color = CurrentTheme.Warning
	}
	return Styled(fmt.Sprintf("%.1f%%", acc), color, Bold)
}

func SparkScore(score float64) string {
	color := CurrentTheme.TextDim
	if score >= 80 {
		color = CurrentTheme.Success
	} else if score >= 50 {
		color = CurrentTheme.Info
	} else if score >= 30 {
		color = CurrentTheme.Warning
	} else if score > 0 {
		color = CurrentTheme.Error
	}
	return Styled(fmt.Sprintf("%.0f", score), color, Bold)
}

func SparkConsistency(con float64) string {
	color := CurrentTheme.Error
	if con >= 90 {
		color = CurrentTheme.Success
	} else if con >= 75 {
		color = CurrentTheme.Info
	} else if con >= 50 {
		color = CurrentTheme.Warning
	}
	return Styled(fmt.Sprintf("%.1f%%", con), color, Bold)
}

type boxKind int

const (
	boxLine boxKind = iota
	boxCenter
	boxSeparator
	boxEmpty
)

type boxEntry struct {
	content string
	kind    boxKind
}

type BoxBuilder struct {
	entries  []boxEntry
	minWidth int
}

func NewBox(minWidth int) *BoxBuilder {
	return &BoxBuilder{minWidth: minWidth}
}

func (b *BoxBuilder) Line(content string) {
	b.entries = append(b.entries, boxEntry{content: content, kind: boxLine})
}

func (b *BoxBuilder) Centered(content string) {
	b.entries = append(b.entries, boxEntry{content: content, kind: boxCenter})
}

func (b *BoxBuilder) Sep() {
	b.entries = append(b.entries, boxEntry{kind: boxSeparator})
}

func (b *BoxBuilder) Blank() {
	b.entries = append(b.entries, boxEntry{kind: boxEmpty})
}

func (b *BoxBuilder) Render() string {
	w := b.minWidth
	for _, e := range b.entries {
		if e.kind == boxLine || e.kind == boxCenter {
			if vl := VisibleLen(e.content); vl > w {
				w = vl
			}
		}
	}

	var s strings.Builder
	s.WriteString("  ╭" + Repeat("─", w) + "╮\n")
	for _, e := range b.entries {
		switch e.kind {
		case boxLine:
			s.WriteString("  │" + PadRight(e.content, w) + "│\n")
		case boxCenter:
			s.WriteString("  │" + Center(e.content, w) + "│\n")
		case boxSeparator:
			s.WriteString("  ├" + Repeat("─", w) + "┤\n")
		case boxEmpty:
			s.WriteString("  │" + Repeat(" ", w) + "│\n")
		}
	}
	s.WriteString("  ╰" + Repeat("─", w) + "╯\n")
	return s.String()
}
