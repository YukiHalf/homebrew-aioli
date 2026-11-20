package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// The full-screen view: the whole page is rendered once, shown on the terminal's alternate screen,
// and stays until q. Scroll with arrows, j/k, space, the mouse wheel. Re-renders on resize.

func termRowsCols() (rows, cols int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 || ws.Row == 0 {
		return 0, 0, false
	}
	return int(ws.Row), int(ws.Col), true
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

type key int

const (
	kNone key = iota
	kQuit
	kUp
	kDown
	kPageUp
	kPageDown
	kTop
	kBottom
	kWheelUp
	kWheelDown
)

type seq struct {
	seq string
	k   key
}

var seqs = []seq{
	{"\x1b[A", kUp}, {"\x1bOA", kUp}, {"\x1b[B", kDown}, {"\x1bOB", kDown},
	{"\x1b[5~", kPageUp}, {"\x1b[6~", kPageDown},
	{"\x1b[H", kTop}, {"\x1b[1~", kTop}, {"\x1bOH", kTop},
	{"\x1b[F", kBottom}, {"\x1b[4~", kBottom}, {"\x1bOF", kBottom},
}

func seqKey(s string) *seq {
	for i := range seqs {
		if strings.HasPrefix(s, seqs[i].seq) {
			return &seqs[i]
		}
	}
	return nil
}

// parseKeys turns a chunk of terminal input into keys.
func parseKeys(b []byte) []key {
	var keys []key
	for i := 0; i < len(b); {
		rest := string(b[i:])
		switch {
		case strings.HasPrefix(rest, "\x1b[<"):
			// SGR mouse: ESC [ < btn ; x ; y (M|m)
			end := strings.IndexAny(rest, "Mm")
			if end < 0 {
				return keys
			}
			btn, _ := strconv.Atoi(strings.SplitN(rest[3:end], ";", 2)[0])
			if btn == 64 {
				keys = append(keys, kWheelUp)
			} else if btn == 65 {
				keys = append(keys, kWheelDown)
			}
			i += end + 1
		case seqKey(rest) != nil:
			s := seqKey(rest)
			keys, i = append(keys, s.k), i+len(s.seq)
		case strings.HasPrefix(rest, "\x1b["):
			// some other sequence: skip to its final byte
			j := 2
			for j < len(rest) && (rest[j] < 0x40 || rest[j] > 0x7e) {
				j++
			}
			i += j + 1
		case b[i] == 0x1b:
			keys, i = append(keys, kQuit), i+1 // a lone Esc
		default:
			switch b[i] {
			case 'q', 'Q', 3, 4: // q, Ctrl-C, Ctrl-D
				keys = append(keys, kQuit)
			case 'k':
				keys = append(keys, kUp)
			case 'j', '\r', '\n':
				keys = append(keys, kDown)
			case ' ', 'f', 6:
				keys = append(keys, kPageDown)
			case 'b', 2:
				keys = append(keys, kPageUp)
			case 'g':
				keys = append(keys, kTop)
			case 'G':
				keys = append(keys, kBottom)
			}
			i++
		}
	}
	return keys
}

// page shows render(width) full screen until the reader quits.
func page(render func(w int) []string, title string) {
	saved, err := stty("-g")
	if err != nil {
		// no controllable terminal: just print
		for _, l := range render(layoutWidth()) {
			fmt.Println(l)
		}
		return
	}
	_, _ = stty("raw", "-echo")
	out := os.Stdout
	restore := func() {
		fmt.Fprint(out, "\x1b[?1006l\x1b[?1000l\x1b[?25h\x1b[?1049l")
		_, _ = stty(saved)
	}
	fmt.Fprint(out, "\x1b[?1049h\x1b[?25l\x1b[?1000h\x1b[?1006h")

	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	input := make(chan []byte)
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				close(input)
				return
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			input <- chunk
		}
	}()

	var lines []string
	var rows, cols, top, margin int
	layout := func() {
		r, c, ok := termRowsCols()
		if !ok {
			r, c = 24, 80
		}
		rows, cols = r, c
		w := layoutWidth()
		lines = render(w)
		margin = (cols - w - 1) / 2
		if margin < 0 {
			margin = 0
		}
	}
	view := func() int { return rows - 1 } // last row is the status bar
	clamp := func() {
		if max := len(lines) - view(); top > max {
			top = max
		}
		if top < 0 {
			top = 0
		}
	}
	draw := func() {
		clamp()
		var b strings.Builder
		b.WriteString("\x1b[H")
		left := strings.Repeat(" ", margin)
		for r := 0; r < view(); r++ {
			b.WriteString("\x1b[" + strconv.Itoa(r+1) + ";1H\x1b[2K")
			if i := top + r; i < len(lines) {
				b.WriteString(left + lines[i])
			}
		}
		// status bar
		pct := "all"
		if len(lines) > view() {
			pct = strconv.Itoa((top+view())*100/len(lines)) + "%"
			if top == 0 {
				pct = "top"
			} else if top+view() >= len(lines) {
				pct = "end"
			}
		}
		help := "↑↓ jk scroll · space page · g/G top/end · q quit"
		leftS := " 🧄 aioli · " + title
		gap := cols - vw(leftS) - vw(help) - vw(pct) - 4
		if gap < 1 {
			help = "q quit"
			gap = cols - vw(leftS) - vw(help) - vw(pct) - 4
		}
		if gap < 1 {
			gap = 1
		}
		bar := leftS + strings.Repeat(" ", gap) + help + "   " + pct + " "
		b.WriteString("\x1b[" + strconv.Itoa(rows) + ";1H\x1b[2K")
		if pal.on {
			b.WriteString(esc + "7;" + fgCode(cOat) + "m" + bar + esc + "0m")
		} else {
			b.WriteString(bar)
		}
		fmt.Fprint(out, b.String())
	}

	layout()
	draw()
	for {
		select {
		case s := <-sigs:
			if s == syscall.SIGWINCH {
				layout()
				draw()
				continue
			}
			restore()
			return
		case chunk, ok := <-input:
			if !ok {
				restore()
				return
			}
			for _, k := range parseKeys(chunk) {
				switch k {
				case kQuit:
					restore()
					return
				case kUp:
					top--
				case kDown:
					top++
				case kWheelUp:
					top -= 3
				case kWheelDown:
					top += 3
				case kPageUp:
					top -= view() - 2
				case kPageDown:
					top += view() - 2
				case kTop:
					top = 0
				case kBottom:
					top = len(lines)
				}
			}
			draw()
		}
	}
}
