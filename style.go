package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unsafe"
)

// ── colour ─────────────────────────────────────────────────────────────────────
// The site's palette (src/styles/theme.css). Truecolor when the terminal says it can, else the nearest of xterm's 256.

const (
	cOat        = "f0e3c2"
	cLavender   = "c6a3dc"
	cLime       = "d5e38c"
	cMustard    = "f1b64c"
	cTangerine  = "f28d52"
	cPeriwinkle = "93a5dc"
	cPink       = "eaa9d4"
	cTomato     = "d9472f"
	cMeadow     = "84b25d"
	cTwilight   = "8f83cc"
	cDawn       = "e9b6d0"
	cPeach      = "f6d4bd"
	cSky        = "c9c4ea"
	cMuted      = "8a8aa6"
)

var moodColor = map[string]string{
	"dizzy": cLavender,
	"derp":  cMustard,
	"meh":   cPeriwinkle,
	"calm":  cLime,
	"baby":  cPink,
	"human": cOat,
}

type palette struct {
	on        bool
	truecolor bool
}

var pal palette

func setupColor(force bool, noColor bool) {
	tty := isTTY()
	pal.on = (tty || force) && !noColor && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
	ct := os.Getenv("COLORTERM")
	pal.truecolor = ct == "truecolor" || ct == "24bit"
}

func rgb(hex string) (int, int, int) {
	v, _ := strconv.ParseUint(hex, 16, 32)
	return int(v >> 16 & 255), int(v >> 8 & 255), int(v & 255)
}

func to256(r, g, b int) int {
	q := func(c int) int {
		if c < 48 {
			return 0
		}
		if c < 115 {
			return 1
		}
		return (c - 35) / 40
	}
	return 16 + 36*q(r) + 6*q(g) + q(b)
}

func fgCode(hex string) string {
	r, g, b := rgb(hex)
	if pal.truecolor {
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	}
	return fmt.Sprintf("38;5;%d", to256(r, g, b))
}

const esc = "\x1b["

// paint wraps s in the given SGR attributes ("1" bold, "2" dim, "3" italic) and colour.
func paint(s, hex string, attrs ...string) string {
	if !pal.on || s == "" {
		return s
	}
	codes := append([]string{}, attrs...)
	if hex != "" {
		codes = append(codes, fgCode(hex))
	}
	if len(codes) == 0 {
		return s
	}
	return esc + strings.Join(codes, ";") + "m" + s + esc + "0m"
}

func bold(s, hex string) string   { return paint(s, hex, "1") }
func dim(s string) string         { return paint(s, cMuted) }
func italic(s, hex string) string { return paint(s, hex, "3") }

// ── terminal ───────────────────────────────────────────────────────────────────

type winsize struct{ Row, Col, X, Y uint16 }

func termSize() (int, bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 0, false
	}
	return int(ws.Col), true
}

func isTTY() bool {
	st, err := os.Stdout.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// width of the whole layout: the terminal, capped so lines stay readable.
func layoutWidth() int {
	w, ok := termSize()
	if !ok {
		if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && c > 0 {
			w = c
		} else {
			w = 80
		}
	}
	if w > 84 {
		w = 84
	}
	if w < 44 {
		w = 44
	}
	return w
}

// ── measuring ──────────────────────────────────────────────────────────────────

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func runeWidth(r rune) int {
	switch {
	case r == 0xAD, r == 0x200B, r == 0x200D, r == 0xFE0F, unicode.Is(unicode.Mn, r):
		return 0
	case r >= 0x1100 && r <= 0x115F, r >= 0x2E80 && r <= 0xA4CF, r >= 0xAC00 && r <= 0xD7A3,
		r >= 0xF900 && r <= 0xFAFF, r >= 0xFE30 && r <= 0xFE4F, r >= 0xFF00 && r <= 0xFF60,
		r >= 0x1F300 && r <= 0x1FAFF:
		return 2
	}
	return 1
}

// visible width of s, ignoring colour codes.
func vw(s string) int {
	n := 0
	for _, r := range ansiRe.ReplaceAllString(s, "") {
		n += runeWidth(r)
	}
	return n
}

func pad(s string, w int) string {
	if d := w - vw(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// ── wrapping ───────────────────────────────────────────────────────────────────

// word is one unbreakable piece of text, already painted.
type word struct {
	s string
	w int
}

// wrapWords lays words into lines no wider than width; a word wider than a line is cut.
func wrapWords(words []word, width int) []string {
	var lines []string
	cur, curW := "", 0
	flush := func() {
		lines = append(lines, cur)
		cur, curW = "", 0
	}
	for _, wd := range words {
		if wd.w > width {
			// hard-cut very long words (URLs), dimmed
			plain := []rune(ansiRe.ReplaceAllString(wd.s, ""))
			for len(plain) > 0 {
				room := width
				if curW > 0 {
					room = width - curW - 1
				}
				if room < 8 {
					flush()
					continue
				}
				n := room
				if n > len(plain) {
					n = len(plain)
				}
				if curW > 0 {
					cur += " "
					curW++
				}
				cur += dim(string(plain[:n]))
				curW += n
				plain = plain[n:]
				if len(plain) > 0 {
					flush()
				}
			}
			continue
		}
		if curW > 0 && curW+1+wd.w > width {
			flush()
		}
		if curW > 0 {
			cur += " "
			curW++
		}
		cur += wd.s
		curW += wd.w
	}
	if curW > 0 {
		lines = append(lines, cur)
	}
	return lines
}

// ── markdown, just enough of it ───────────────────────────────────────────────

var (
	linkRe    = regexp.MustCompile(`!?\[([^\]]*)\]\(([^)\s]+)[^)]*\)`)
	listRe    = regexp.MustCompile(`^(\s*)([-*+]|\d+[.)])\s+(.*)$`)
	headRe    = regexp.MustCompile(`^#{1,6}\s+(.*)$`)
	quoteRe   = regexp.MustCompile(`^>\s?(.*)$`)
	emphTokRe = regexp.MustCompile("(\\*\\*|__|\\*|_|`)")
)

const softHyphen = "\xc2\xad"

// inline turns a markdown line into painted words: **bold**, *italic*, `code`, [links](url).
func inline(text, color string, baseAttrs ...string) []word {
	text = strings.ReplaceAll(text, softHyphen, "")
	text = linkRe.ReplaceAllStringFunc(text, func(m string) string {
		sub := linkRe.FindStringSubmatch(m)
		if strings.HasPrefix(m, "!") {
			return "[image: " + sub[1] + "]"
		}
		if sub[1] == sub[2] {
			return sub[2]
		}
		return sub[1] + " ⟨" + sub[2] + "⟩"
	})
	var words []word
	boldOn, italOn, codeOn := false, false, false
	for _, raw := range strings.Fields(text) {
		var b strings.Builder
		w := 0
		parts := emphTokRe.Split(raw, -1)
		marks := emphTokRe.FindAllString(raw, -1)
		for i, p := range parts {
			if p != "" {
				attrs := append([]string{}, baseAttrs...)
				c := color
				if boldOn {
					attrs = append(attrs, "1")
				}
				if italOn {
					attrs = append(attrs, "3")
				}
				if codeOn {
					c = cLime
				}
				if strings.HasPrefix(p, "⟨") || strings.HasSuffix(p, "⟩") {
					b.WriteString(dim(p))
				} else {
					b.WriteString(paint(p, c, attrs...))
				}
				w += vw(p)
			}
			if i < len(marks) {
				switch marks[i] {
				case "**", "__":
					boldOn = !boldOn
				case "*":
					italOn = !italOn
				case "_":
					// _ is emphasis only at a word's edge, so snake_case stays
					if i == 0 && p == "" || i == len(marks)-1 && parts[i+1] == "" {
						italOn = !italOn
					} else {
						b.WriteString(paint("_", color, baseAttrs...))
						w++
					}
				case "`":
					codeOn = !codeOn
				}
			}
		}
		if w > 0 {
			words = append(words, word{b.String(), w})
		}
	}
	return words
}

// renderMarkdown wraps a markdown body to width, keeping paragraphs, lists, headings and quotes.
func renderMarkdown(body string, width int, color string, attrs ...string) []string {
	var out []string
	var para []string
	flushPara := func() {
		if len(para) == 0 {
			return
		}
		out = append(out, wrapWords(inline(strings.Join(para, " "), color, attrs...), width)...)
		para = nil
	}
	blank := func() {
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
	}
	inFence := false
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		t := strings.TrimRight(line, " \t")
		if strings.HasPrefix(strings.TrimSpace(t), "```") {
			flushPara()
			inFence = !inFence
			continue
		}
		if inFence {
			out = append(out, paint(t, cLime))
			continue
		}
		switch {
		case strings.TrimSpace(t) == "":
			flushPara()
			blank()
		case headRe.MatchString(t):
			flushPara()
			blank()
			out = append(out, wrapWords(inline(headRe.FindStringSubmatch(t)[1], cTangerine, "1"), width)...)
		case listRe.MatchString(t):
			flushPara()
			m := listRe.FindStringSubmatch(t)
			bullet := "•"
			if !strings.ContainsAny(m[2], "-*+") {
				bullet = m[2]
			}
			head := strings.Repeat(" ", len(m[1])) + paint(bullet, cTangerine) + " "
			hang := strings.Repeat(" ", vw(head))
			for i, l := range wrapWords(inline(m[3], color, attrs...), width-vw(head)) {
				if i == 0 {
					out = append(out, head+l)
				} else {
					out = append(out, hang+l)
				}
			}
		case quoteRe.MatchString(t):
			flushPara()
			for _, l := range wrapWords(inline(quoteRe.FindStringSubmatch(t)[1], color, "3"), width-2) {
				out = append(out, paint("▎", cTwilight)+" "+l)
			}
		default:
			para = append(para, strings.TrimSpace(t))
		}
	}
	flushPara()
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

// ── boxes ──────────────────────────────────────────────────────────────────────

// box draws a rounded box of outer width w around content lines; labels go into the top and bottom edges.
func box(lines []string, w int, color, header, footer string) []string {
	inner := w - 4
	edge := func(l, r, label string) string {
		if label == "" {
			return paint(l+strings.Repeat("─", w-2)+r, color)
		}
		fill := w - 2 - 2 - vw(label) - 1
		if fill < 1 {
			fill = 1
		}
		return paint(l+"─ ", color) + label + paint(" "+strings.Repeat("─", fill)+r, color)
	}
	out := []string{edge("╭", "╮", header)}
	for _, l := range lines {
		out = append(out, paint("│", color)+" "+pad(l, inner)+" "+paint("│", color))
	}
	return append(out, edge("╰", "╯", footer))
}

// sideBySide puts art left of block, top-aligned.
func sideBySide(art []string, artW int, block []string) []string {
	n := len(block)
	if len(art) > n {
		n = len(art)
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		a, b := "", ""
		if i < len(art) {
			a = art[i]
		}
		if i < len(block) {
			b = block[i]
		}
		out[i] = pad(a, artW) + b
	}
	return out
}

func center(s string, w int) string {
	d := (w - vw(s)) / 2
	if d < 0 {
		d = 0
	}
	return strings.Repeat(" ", d) + s
}
