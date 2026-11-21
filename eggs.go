package main

import (
	"math/rand"
	"strings"
	"time"
)

// ── easter egg 1: make the AI dumber ───────────────────────────────────────────
// d writes one more version of the post with the tiniest model of all: a character-level Markov chain,
// trained on the spot on every aioli post. Each press drops its order by one, so it gets dumber.

const maxDumb = 6

func markovVersion(f *Feed, lang string, level int, seed int64) string {
	order := maxDumb - 1 - level // level 1 → order 4 … level 5 → order 0
	if level >= maxDumb {
		return ""
	}
	var sb strings.Builder
	for _, p := range f.Posts {
		if t, ok := p.Human[lang]; ok {
			sb.WriteString(t.Body + " ")
		}
		for _, v := range p.Ai {
			if v.Lang == lang {
				sb.WriteString(v.Body + " ")
			}
		}
	}
	text := []rune(strings.Join(strings.Fields(strings.NewReplacer("*", "", "#", "", "_", "", "`", "").Replace(sb.String())), " "))
	if len(text) < 50 {
		return "..."
	}
	rng := rand.New(rand.NewSource(seed))
	if order <= 0 {
		var out []rune
		for i := 0; i < 220; i++ {
			out = append(out, text[rng.Intn(len(text))])
		}
		return string(out)
	}
	next := map[string][]rune{}
	for i := 0; i+order < len(text); i++ {
		k := string(text[i : i+order])
		next[k] = append(next[k], text[i+order])
	}
	// start at the beginning of a sentence
	start := 0
	for tries := 0; tries < 50; tries++ {
		i := rng.Intn(len(text) - order)
		if i == 0 || text[i-1] == ' ' && i > 1 && strings.ContainsRune(".!?", text[i-2]) {
			start = i
			break
		}
	}
	out := append([]rune{}, text[start:start+order]...)
	for len(out) < 600 {
		opts := next[string(out[len(out)-order:])]
		if len(opts) == 0 {
			break
		}
		r := opts[rng.Intn(len(opts))]
		out = append(out, r)
		if len(out) > 260 && strings.ContainsRune(".!?", r) {
			break
		}
	}
	return string(out)
}

var dumbWords = map[string][]string{
	"en": {
		"",
		"order 4 · almost words",
		"order 3 · confident",
		"order 2 · it has stopped trying",
		"order 1 · vowels, mostly",
		"order 0 · just letters now",
		"it can't get dumber. it is a garlic now.",
	},
	"de": {
		"",
		"Ordnung 4 · fast Wörter",
		"Ordnung 3 · selbstbewusst",
		"Ordnung 2 · hat aufgegeben",
		"Ordnung 1 · hauptsächlich Vokale",
		"Ordnung 0 · nur noch Buchstaben",
		"dümmer geht nicht. es ist jetzt eine Knoblauchzehe.",
	},
}

// markovLine: where the dumb card starts in the rendered page, so the view can jump to it.
var markovLine = -1

func renderDumber(st *state, w int) {
	markovLine = len(buf)
	cw := w - 3 - 8
	c := cLime
	level := st.dumber
	if level > maxDumb {
		level = maxDumb
	}
	face := []string{"(o_o)", "(o_o)", "(O_o)", "(@_@)", "(x_@)", "(._.)", "( )"}[level]
	head := bold("0.000001b", c) + dim(" · markov chain, "+map[string]string{"en": "trained just now", "de": "gerade eben trainiert"}[st.lang])
	lines := []string{head, ""}
	if level >= maxDumb {
		lines = append(lines, renderGarlicMini()...)
	} else {
		shown, _ := shownLang(st.f.Posts[st.post], st.lang)
		lines = append(lines, renderMarkdown(markovVersion(st.f, shown, level, st.seed), cw-4, cCream)...)
	}
	press := map[string]string{"en": "d: dumber", "de": "d: dümmer"}[st.lang]
	foot := dim(dumbWords[st.lang][level])
	if level < maxDumb {
		foot += dim(" · " + press)
	}
	printLines(indent(card(lines, cw, c, []string{chip(face, cSpace, c)}, foot), 1+8))
	nl()
}

func renderGarlicMini() []string {
	g := garlicCells(4)
	var out []string
	for _, row := range g {
		line := ""
		for _, cell := range row {
			if cell.ch == ' ' {
				line += " "
			} else {
				line += paint(string(cell.ch), toneColor[cell.tone])
			}
		}
		out = append(out, "  "+line)
	}
	return out
}

// ── easter egg 2: the Konami code makes the garlic laugh ──────────────────────

var konami = []int{kUp, kUp, kDown, kDown, kLeft, kRight, kLeft, kRight, kPageUp, 'a'} // b arrives as page-up

// konamiStep feeds one key; swallow = don't act on it (it's part of the code), done = the code is complete.
func (st *state) konamiStep(k int) (swallow, done bool) {
	if k == konami[st.konamiPos] {
		st.konamiPos++
		if st.konamiPos == len(konami) {
			st.konamiPos = 0
			return true, true
		}
		return st.konamiPos > 4, false
	}
	st.konamiPos = 0
	if k == konami[0] {
		st.konamiPos = 1
	}
	return false, false
}

// ── easter egg 3: sudo ─────────────────────────────────────────────────────────

func sudo(lang string) {
	if lang == "de" {
		println("aioli: netter Versuch. Knoblauch hat keine Root-Rechte. Nur Wurzeln.")
	} else {
		println("aioli: nice try. garlic doesn't have root access. it only has roots.")
	}
}

func newSeed() int64 { return time.Now().UnixNano() }
