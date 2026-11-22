package main

import (
	"fmt"
	"math"
	"strings"
)

// The look of the site's ASCII home (branch site-pages-canvas): a shaded garlic in space, roots that turn
// into the navigation, a prompt line, the "ruined" headline, outlined cards with face stickers and a
// striped shadow. Colours are the site's.

const (
	cSpace = "1c1a46"
	cCream = "fbf4de"
	cNight = "35337f"
)

// sticker colour + face per model, as on the home cards
var modelFace = map[string]struct{ face, color string }{
	"smol-135m":      {"(x_x)", cTangerine},
	"qwen-0.5b":      {"(o_O)", cMustard},
	"tinyllama-1.1b": {"(-_-)", cPeriwinkle},
	"gabrijel":       {"(^ᴗ^)", cPink},
}

func faceOf(model string) (string, string) {
	if f, ok := modelFace[model]; ok {
		return f.face, f.color
	}
	return "(?_?)", cLavender
}

var cardPalette = []string{cLime, cLavender, cMustard, cPeriwinkle}

// ── the garlic ─────────────────────────────────────────────────────────────────
// src/lib/ascii-garlic.ts: a surface of revolution lit from the top left, clove ridges, a face cut out.
// Terminal cells are about twice as tall as wide, hence rx = 2.1 × ry.

type gcell struct {
	ch   byte
	tone string
}

var toneColor = map[string]string{"hi": cCream, "mid": cPink, "lo": cLavender, "deep": cPeriwinkle}

func garlicCells(ry float64) [][]gcell {
	R := int(2*ry) - 1
	rx := ry * 2.1
	C := int(math.Ceil(rx*2)) + 2
	cx := float64(C) / 2
	cy := float64(R)/2 - 0.5
	ramp := " .:-=+*#%@"
	L := []float64{-0.5, -0.55, 0.67}
	lm := math.Sqrt(L[0]*L[0] + L[1]*L[1] + L[2]*L[2])
	for i := range L {
		L[i] /= lm
	}
	halfW := func(y float64) float64 {
		if y < -1 || y > 0.95 {
			return 0
		}
		if y < -0.15 {
			return 0.66 * math.Pow((y+1)/0.85, 1.7)
		}
		return 0.92 * math.Sqrt(math.Max(0, 1-math.Pow((y-0.28)/0.67, 2)))
	}
	// the face has to survive small sizes: at least about a cell wide / tall
	eyeW, eyeH := math.Max(0.1, 1.1/rx), math.Max(0.13, 0.7/ry)
	smile := math.Max(0.05, 0.5/ry)
	out := make([][]gcell, R)
	for r := 0; r < R; r++ {
		row := make([]gcell, C)
		for c := 0; c < C; c++ {
			X := (float64(c) + 0.5 - cx) / rx
			Y := (float64(r) + 0.5 - cy) / ry
			w := halfW(Y)
			cell := gcell{' ', ""}
			if w > 0 && math.Abs(X) < w {
				nx := X / w
				nz := math.Sqrt(1 - nx*nx)
				slope := (halfW(Y+0.02) - halfW(Y-0.02)) / 0.04
				n := []float64{nx, -slope * 0.5, nz}
				m := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
				d := math.Max(0, (n[0]*L[0]+n[1]*L[1]+n[2]*L[2])/m)
				if math.Abs(math.Abs(nx)-0.42) < 0.05 && Y > -0.5 {
					d *= 0.35 // clove ridges
				}
				i := 1 + int(d*float64(len(ramp)-1))
				if i > len(ramp)-1 {
					i = len(ramp) - 1
				}
				cell.ch = ramp[i]
				switch {
				case d > 0.78:
					cell.tone = "hi"
				case d > 0.5:
					cell.tone = "mid"
				case d > 0.25:
					cell.tone = "lo"
				default:
					cell.tone = "deep"
				}
				for _, ex := range []float64{-0.3, 0.3} {
					if math.Pow((X-ex)/eyeW, 2)+math.Pow((Y-0.3)/eyeH, 2) < 1 {
						cell = gcell{' ', ""}
					}
				}
				if math.Abs(X) < 0.24 && math.Abs(Y-(0.62-1.6*X*X)) < smile {
					cell = gcell{' ', ""}
				}
			}
			row[c] = cell
		}
		out[r] = row
	}
	return out
}

// ── hero: rings, garlic, roots, nav, prompt, headline ─────────────────────────

type navItem struct {
	key, label, color string
	x                 float64 // -1 … 1: where its root ends
	brand             bool
}

func hero(st *state, w int, rows int) []string {
	s := dict[st.lang]
	// garlic size from the terminal height: the hero should fit on the first screen
	ry := float64(rows-14) / 2.3
	ry = math.Max(5, math.Min(10, ry))
	g := garlicCells(ry)
	laughRow := -1
	if st.laugh {
		// the site's laugh: the garlic splits at its mouth and the top half hinges open
		split := int(math.Round(float64(len(g))/2 - 0.5 + 0.62*ry - 0.5))
		blank := make([]gcell, len(g[0]))
		for i := range blank {
			blank[i] = gcell{' ', ""}
		}
		top := g[:split]
		// tilt the top half: each row further up shifts a little more to the left
		tilted := make([][]gcell, len(top))
		for i, row := range top {
			shift := (len(top) - i) / 3
			nr := make([]gcell, len(row))
			for c := range nr {
				if c+shift < len(row) {
					nr[c] = row[c+shift]
				} else {
					nr[c] = gcell{' ', ""}
				}
			}
			tilted[i] = nr
		}
		g = append(append(append([][]gcell{}, tilted...), blank, blank), g[split:]...)
		laughRow = split
	}
	R, C := len(g), len(g[0])

	H := R + 2
	canvas := make([][]string, H) // one painted string per cell
	for i := range canvas {
		canvas[i] = make([]string, w)
		for j := range canvas[i] {
			canvas[i][j] = " "
		}
	}
	gx, gy := (w-C)/2, 1
	ccx, ccy := float64(gx)+float64(C)/2, float64(gy)+float64(R)/2-0.5

	// the rings: thin ellipses round the garlic in the sticker colours, fading outward, clipped by the stage
	ringCols := []string{cLavender, cMustard, cLavender, cTangerine, cPeriwinkle, cMustard}
	for k := 0; k < len(ringCols); k++ {
		a := float64(C)/2 + 2.5 + float64(k)*3.2
		b := a / 2.1
		col := mix(ringCols[k], cSpace, 0.12+float64(k)*0.11)
		steps := int(a * 12)
		for i := 0; i < steps; i++ {
			t := 2 * math.Pi * float64(i) / float64(steps)
			x := int(math.Floor(ccx + a*math.Cos(t)))
			y := int(math.Floor(ccy + b*math.Sin(t)))
			if y < 0 || y >= H || x < 0 || x >= w {
				continue
			}
			// the line's direction picks the character
			dx, dy := -a*math.Sin(t), b*math.Cos(t)*2.1
			ang := math.Atan2(dy, dx)
			ch := "|"
			switch s := math.Abs(math.Tan(ang)); {
			case s < 0.45:
				ch = "-"
			case s < 2.2:
				if math.Tan(ang) < 0 {
					ch = "/"
				} else {
					ch = "\\"
				}
			}
			canvas[y][x] = paint(ch, col)
		}
	}
	// a dark disc behind the garlic, like its round backdrop
	for y := 0; y < H; y++ {
		for x := 0; x < w; x++ {
			dx, dy := (float64(x)+0.5-ccx)/(float64(C)/2+1.5), (float64(y)+0.5-ccy)/((float64(C)/2+1.5)/2.1)
			if dx*dx+dy*dy < 1 {
				canvas[y][x] = " "
			}
		}
	}
	for r, row := range g {
		for c, cell := range row {
			if cell.ch != ' ' {
				canvas[gy+r][gx+c] = paint(string(cell.ch), toneColor[cell.tone])
			}
		}
	}
	if laughRow >= 0 {
		ha := "A H A H A H A H A"
		x0 := gx + (C-len(ha))/2
		for i, ch := range ha {
			if ch != ' ' {
				canvas[gy+laughRow][x0+i] = bold(string(ch), cTangerine)
			}
		}
	}
	out := make([]string, 0, H+12)
	for _, row := range canvas {
		out = append(out, strings.Join(row, ""))
	}

	// roots from the bottom of the bulb out to the five nav items
	other := "de"
	if st.lang == "de" {
		other = "en"
	}
	nav := []navItem{
		{"today", s.navToday, cLime, -1, false},
		{"list", s.navPosts, cMustard, -0.5, false},
		{"brand", "aioli", cTangerine, 0, true},
		{"hof", s.navFame, cPeriwinkle, 0.5, false},
		{"lang", strings.ToUpper(other), cLavender, 1, false},
	}
	spread := float64(w)/2 - 7
	RR := 5
	roots := make([][]string, RR)
	for i := range roots {
		roots[i] = make([]string, w)
		for j := range roots[i] {
			roots[i][j] = " "
		}
	}
	mid := float64(w) / 2
	origins := []float64{-0.16, -0.08, 0, 0.08, 0.16}
	ends := make([]int, len(nav))
	for i, n := range nav {
		c0 := mid + origins[i]*ry*2.1
		c1 := mid + n.x*spread
		prev := int(math.Round(c0))
		for r := 0; r < RR; r++ {
			t := float64(r) / float64(RR-1)
			e := t * t * (3 - 2*t)
			col := int(math.Round(c0 + (c1-c0)*e))
			lo, hi := prev, col
			if lo > hi {
				lo, hi = hi, lo
			}
			for c := lo; c <= hi; c++ {
				if c != col && c >= 0 && c < w {
					roots[r][c] = paint("_", n.color)
				}
			}
			ch := "|"
			if r == RR-1 {
				ch = "o"
			} else if col > prev {
				ch = "\\"
			} else if col < prev {
				ch = "/"
			}
			if col >= 0 && col < w {
				roots[r][col] = paint(ch, n.color)
			}
			prev = col
		}
		ends[i] = prev
	}
	for _, row := range roots {
		out = append(out, strings.Join(row, ""))
	}
	// the nav row: each label centred under its root
	line := ""
	used := 0
	for i, n := range nav {
		label := "[ " + n.label + " ]"
		var shown string
		switch {
		case n.brand:
			shown = bold(n.label, cTangerine)
			label = n.label
		case n.key == st.mode || (n.key == "today" && st.mode == "day" && st.isToday):
			shown = chip("  "+n.label+"  ", cSpace, n.color)
		default:
			shown = paint("[ ", n.color) + paint(n.label, n.color, "1") + paint(" ]", n.color)
		}
		start := ends[i] - vw(label)/2
		if start < used+1 && i > 0 {
			start = used + 1
		}
		if start < 0 {
			start = 0
		}
		line += strings.Repeat(" ", start-used) + shown
		used = start + vw(label)
	}
	out = append(out, line, "")

	// prompt + headline
	out = append(out, "  "+paint("> ", cLime)+paint(st.command(), cLime)+paint("█", cLime))
	out = append(out, "")
	out = append(out, bigHeadline(s.hero, w-4)...)
	out = append(out, "  "+paint("[ "+s.heroNote+" ]", cMustard), "")
	return out
}

// ── cards ──────────────────────────────────────────────────────────────────────

// card draws an outlined card of outer width w: stickers sit in the top edge on the right,
// a striped shadow falls to the right and below.
func card(lines []string, w int, color string, stickers []string, bottom string) []string {
	inner := w - 4
	top := "┌"
	stick := strings.Join(stickers, " ")
	if stick != "" {
		fill := w - vw(stick) - 5
		top += strings.Repeat("─", fill) + " "
		top = paint(top, color) + stick + paint(" ─┐", color)
	} else {
		top = paint(top+strings.Repeat("─", w-2)+"┐", color)
	}
	shade := mix(color, cSpace, 0.55)
	shadow := paint("░", shade)
	out := []string{top + " "}
	for _, l := range lines {
		out = append(out, paint("│", color)+" "+pad(l, inner)+" "+paint("│", color)+shadow)
	}
	if bottom != "" {
		fill := w - 2 - 2 - vw(bottom) - 1
		if fill < 1 {
			fill = 1
		}
		out = append(out, paint("└─ ", color)+bottom+paint(" "+strings.Repeat("─", fill)+"┘", color)+shadow)
	} else {
		out = append(out, paint("└"+strings.Repeat("─", w-2)+"┘", color)+shadow)
	}
	out = append(out, " "+paint(strings.Repeat("░", w), shade))
	return out
}

// indent shifts a block right; the cards are staggered like on the home page.
func indent(block []string, n int) []string {
	if n <= 0 {
		return block
	}
	sp := strings.Repeat(" ", n)
	out := make([]string, len(block))
	for i, l := range block {
		out[i] = sp + l
	}
	return out
}

func staggerOf(i, room int) int {
	switch i % 4 {
	case 0:
		return 0
	case 1:
		return room
	case 2:
		return room / 3
	default:
		return room * 2 / 3
	}
}

// ── footer ─────────────────────────────────────────────────────────────────────

var footerArt = []string{
	"           _         _  _ ",
	"    __ _  (_)  ___  | |(_)",
	"   / _` | | | / _ \\ | || |",
	"  | (_| | | || (_) || || |",
	"   \\__,_| |_| \\___/ |_||_|",
}

func footer(lang string, w int) []string {
	out := []string{"", paint(strings.Repeat("─", w), cNight), ""}
	for _, l := range footerArt {
		out = append(out, center(paint(l, cLime), w))
	}
	out = append(out, "", center(paint(dict[lang].tagline, cLime), w), "")
	return out
}

// mix blends colour a toward b by t (0 = a, 1 = b).
func mix(a, b string, t float64) string {
	ar, ag, ab := rgb(a)
	br, bg, bb := rgb(b)
	m := func(x, y int) int { return int(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return fmt.Sprintf("%02x%02x%02x", m(ar, br), m(ag, bg), m(ab, bb))
}

// ── the headline in pixel letters ──────────────────────────────────────────────
// A 3×5 pixel font drawn with half blocks, so each letter is 3 cells wide and 3 rows tall,
// standing in for the site's Pixelify Sans.

var pixelFont = map[rune][5]string{
	'a': {".#.", "#.#", "###", "#.#", "#.#"}, 'b': {"##.", "#.#", "##.", "#.#", "##."},
	'c': {".##", "#..", "#..", "#..", ".##"}, 'd': {"##.", "#.#", "#.#", "#.#", "##."},
	'e': {"###", "#..", "##.", "#..", "###"}, 'f': {"###", "#..", "##.", "#..", "#.."},
	'g': {".##", "#..", "#.#", "#.#", ".##"}, 'h': {"#.#", "#.#", "###", "#.#", "#.#"},
	'i': {"###", ".#.", ".#.", ".#.", "###"}, 'j': {"..#", "..#", "..#", "#.#", ".#."},
	'k': {"#.#", "#.#", "##.", "#.#", "#.#"}, 'l': {"#..", "#..", "#..", "#..", "###"},
	'm': {"#.#", "###", "###", "#.#", "#.#"}, 'n': {"##.", "#.#", "#.#", "#.#", "#.#"},
	'o': {".#.", "#.#", "#.#", "#.#", ".#."}, 'p': {"##.", "#.#", "##.", "#..", "#.."},
	'q': {".#.", "#.#", "#.#", "##.", ".##"}, 'r': {"##.", "#.#", "##.", "#.#", "#.#"},
	's': {".##", "#..", ".#.", "..#", "##."}, 't': {"###", ".#.", ".#.", ".#.", ".#."},
	'u': {"#.#", "#.#", "#.#", "#.#", "###"}, 'v': {"#.#", "#.#", "#.#", "#.#", ".#."},
	'w': {"#.#", "#.#", "###", "###", "#.#"}, 'x': {"#.#", "#.#", ".#.", "#.#", "#.#"},
	'y': {"#.#", "#.#", ".#.", ".#.", ".#."}, 'z': {"###", "..#", ".#.", "#..", "###"},
	'ä': {"#.#", "...", "###", "#.#", "#.#"}, 'ö': {"#.#", "...", "###", "#.#", "###"},
	'ü': {"#.#", "...", "#.#", "#.#", "###"}, 'ß': {"##.", "#.#", "##.", "#.#", "##."},
	',': {"...", "...", "...", ".#.", "#.."}, '.': {"...", "...", "...", "...", ".#."},
	'-': {"...", "...", "###", "...", "..."}, '!': {".#.", ".#.", ".#.", "...", ".#."},
	' ': {"..", "..", "..", "..", ".."},
}

// pixelWord draws one word as 3 rows of half blocks (the word includes a trailing gap column per letter).
func pixelWord(s string) [3]string {
	var rows [3]strings.Builder
	for _, r := range strings.ToLower(s) {
		g, ok := pixelFont[r]
		if !ok {
			g = pixelFont['-']
		}
		for x := 0; x < len(g[0]); x++ {
			on := func(y int) bool { return y < 5 && g[y][x] == '#' }
			for i := 0; i < 3; i++ {
				top, bot := on(2*i), on(2*i+1)
				switch {
				case top && bot:
					rows[i].WriteString("█")
				case top:
					rows[i].WriteString("▀")
				case bot:
					rows[i].WriteString("▄")
				default:
					rows[i].WriteString(" ")
				}
			}
		}
		for i := 0; i < 3; i++ {
			rows[i].WriteString(" ")
		}
	}
	return [3]string{rows[0].String(), rows[1].String(), rows[2].String()}
}

// bigHeadline sets the slogan in pixel letters, wrapping by word, the middle part on a tangerine block.
func bigHeadline(parts [3]string, w int) []string {
	type piece struct {
		rows  [3]string
		width int
		mark  bool
		br    bool
	}
	var pieces []piece
	for i, part := range parts {
		for _, word := range strings.Fields(part) {
			r := pixelWord(word)
			pieces = append(pieces, piece{rows: r, width: vw(r[0]), mark: i == 1})
		}
		if i == 0 {
			pieces = append(pieces, piece{br: true}) // the site breaks after the first part
		}
	}
	var out []string
	var line [3]string
	used := 0
	flush := func() {
		if used == 0 {
			return
		}
		for _, l := range line {
			out = append(out, "  "+l)
		}
		out = append(out, "")
		line, used = [3]string{}, 0
	}
	for _, p := range pieces {
		if p.br {
			flush()
			continue
		}
		pw := p.width + 1
		if p.mark {
			pw += 2
		}
		if used > 0 && used+pw > w {
			flush()
		}
		for i := 0; i < 3; i++ {
			if p.mark {
				line[i] += chip(" "+p.rows[i]+" ", cSpace, cTangerine) + " "
			} else {
				line[i] += bold(p.rows[i], cCream) + " "
			}
		}
		used += pw
	}
	flush()
	return out
}
