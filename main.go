// aioli: daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var version = "dev"

// ── words, in both languages ───────────────────────────────────────────────────

type words struct {
	tagline, human, writtenAs, onlyIn, aiWarning, gabTeaser, nothingToday, latestInstead,
	noMatch, alsoIn, ruinedBy, hofIntro, from, roughly, offline, readMore, listHint string
	months []string
}

var dict = map[string]words{
	"en": {
		tagline:       "daily lab notes, ruined by tiny AIs",
		human:         "human original",
		writtenAs:     "written as",
		onlyIn:        "only in %s, sorry",
		aiWarning:     "written by very small AIs. nothing below is true.",
		gabTeaser:     "gabrijel is building his own tiny AI from scratch. it can't write yet. soon.",
		nothingToday:  "nothing cooked today yet. the lab is either working or asleep.",
		latestInstead: "the latest mess instead",
		noMatch:       "no post matches %q. try: aioli --list",
		alsoIn:        "also in",
		ruinedBy:      "then the tiny AIs got it",
		hofIntro:      "the best things our tiny models ever said, hand-picked.",
		from:          "from",
		roughly:       "roughly:",
		offline:       "offline: this is the copy saved %s",
		readMore:      "on the site",
		listHint:      "read one: aioli <slug or a word from the title>",
		months:        []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	},
	"de": {
		tagline:       "tägliche Lab-Notizen, ruiniert von winzigen KIs",
		human:         "das Original vom Menschen",
		writtenAs:     "geschrieben als",
		onlyIn:        "nur auf %s, sorry",
		aiWarning:     "geschrieben von sehr kleinen KIs. nichts hiervon stimmt.",
		gabTeaser:     "gabrijel baut gerade seine eigene winzige KI von Grund auf. schreiben kann sie noch nicht. bald.",
		nothingToday:  "heute noch nichts gekocht. das Lab arbeitet oder schläft.",
		latestInstead: "dafür das neueste Chaos",
		noMatch:       "kein Post passt zu %q. probier: aioli --list",
		alsoIn:        "auch auf",
		ruinedBy:      "dann kamen die winzigen KIs",
		hofIntro:      "das Beste, was unsere winzigen Modelle je gesagt haben, handverlesen.",
		from:          "aus",
		roughly:       "ungefähr:",
		offline:       "offline: das ist die Kopie vom %s",
		readMore:      "auf der Seite",
		listHint:      "lesen: aioli <slug oder ein Wort aus dem Titel>",
		months:        []string{"Jan.", "Feb.", "März", "Apr.", "Mai", "Juni", "Juli", "Aug.", "Sep.", "Okt.", "Nov.", "Dez."},
	},
}

func langName(l, in string) string {
	names := map[string]map[string]string{
		"en": {"en": "english", "de": "german"},
		"de": {"en": "englisch", "de": "deutsch"},
	}
	return names[in][l]
}

func fmtDate(iso, lang string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	m := dict[lang].months[t.Month()-1]
	if lang == "de" {
		return fmt.Sprintf("%d. %s %d", t.Day(), m, t.Year())
	}
	return fmt.Sprintf("%d %s %d", t.Day(), m, t.Year())
}

// ── art ────────────────────────────────────────────────────────────────────────

var bannerArt = []string{
	"             _         _  _ ",
	"      __ _  (_)  ___  | |(_)",
	"     / _` | | | / _ \\ | || |",
	"    | (_| | | || (_) || || |",
	"     \\__,_| |_| \\___/ |_||_|",
}

var bannerColors = []string{cSky, cTwilight, cDawn, cPeach, cMustard}

func banner(w int, lang string) []string {
	out := []string{""}
	for i, l := range bannerArt {
		out = append(out, bold(l, bannerColors[i]))
	}
	out = append(out, "    "+paint("🧄 "+dict[lang].tagline, cOat, "3"), "")
	_ = w
	return out
}

type face struct{ l, r, mouth string }

var faces = map[string]face{
	"dizzy": {"@", "@", "~"},
	"derp":  {"o", "O", "▽"},
	"meh":   {"-", "-", "_"},
	"calm":  {"^", "^", "‿"},
	"baby":  {"•", "•", "o"},
}

const artW = 13

// garlic draws the aioli clove in a mood.
func garlic(mood string) []string {
	f, ok := faces[mood]
	if !ok {
		f = faces["derp"]
	}
	c := moodColor[mood]
	if c == "" {
		c = cMustard
	}
	o := func(s string) string { return paint(s, c) }
	e := func(s string) string { return bold(s, cOat) }
	return []string{
		"    " + paint(",|,", cMeadow),
		"  " + o(".'") + paint("/|\\", cMeadow) + o("'."),
		" " + o("/ ") + e(f.l) + "   " + e(f.r) + o(" \\"),
		o("|") + paint(" .", cPink) + "  " + e(f.mouth) + "  " + paint(". ", cPink) + o("|"),
		" " + o("'._______.'"),
	}
}

// bubble is a box with a tail pointing left, at the garlic's face.
func bubble(lines []string, w int, color, header, footer string) []string {
	b := box(lines, w, color, header, footer)
	if len(b) > 3 {
		b[2] = paint("◂", color) + strings.TrimPrefix(b[2], paint("│", color))
	}
	return b
}

// ── views ──────────────────────────────────────────────────────────────────────

type opts struct {
	lang      string
	model     string
	humanOnly bool
	aiOnly    bool
	open      bool
}

func printLines(ls []string) {
	for _, l := range ls {
		fmt.Println(" " + strings.TrimRight(l, " "))
	}
}

// renderPost prints one post: the human original, then every tiny AI's version of it.
func renderPost(f *Feed, p Post, o opts, w int) {
	s := dict[o.lang]
	shown := o.lang
	text, ok := p.Human[shown]
	if !ok {
		for l, t := range p.Human {
			shown, text = l, t
			break
		}
	}

	if !o.aiOnly {
		meta := paint(fmtDate(p.Date, o.lang), cPeach) + dim(" · ") + paint(p.Author, cOat) + dim(" · "+shown)
		body := wrapWords(inline(text.Title, cTangerine, "1"), w-4)
		body = append(body, "")
		body = append(body, renderMarkdown(text.Body, w-4, "")...)
		foot := dim(s.human)
		if shown != o.lang {
			foot = paint(fmt.Sprintf(s.onlyIn, langName(shown, o.lang)), cTomato)
		} else if len(p.Human) > 1 {
			var others []string
			for l := range p.Human {
				if l != shown {
					others = append(others, l)
				}
			}
			foot += dim(" · " + s.alsoIn + " " + strings.Join(others, ", ") + " (--lang " + others[0] + ")")
		}
		printLines(box(body, w, cOat, meta, foot))
	}
	if o.humanOnly {
		fmt.Println()
		return
	}

	var versions []AiVersion
	for _, v := range p.Ai {
		if v.Lang != shown {
			continue
		}
		if o.model != "" {
			m := f.model(v.Model)
			q := strings.ToLower(o.model)
			if !strings.Contains(strings.ToLower(m.ID+" "+m.Label+" "+m.Name), q) {
				continue
			}
		}
		versions = append(versions, v)
	}
	if len(versions) > 0 {
		if !o.aiOnly {
			fmt.Println()
			printLines([]string{"     " + paint("↓ ", cTangerine) + paint(s.ruinedBy, cMuted, "3") + paint(" ↓", cTangerine)})
		}
		printLines([]string{"     " + paint("⚠ "+s.aiWarning, cTomato, "3")})
		fmt.Println()
		for _, v := range versions {
			m := f.model(v.Model)
			c := moodColor[m.Mood]
			head := bold(m.Label, c) + dim(" · "+m.Name)
			foot := dim(s.writtenAs + " " + f.persona(v.Persona, o.lang))
			b := bubble(renderMarkdown(v.Body, w-artW-4, ""), w-artW, c, head, foot)
			printLines(sideBySide(garlic(m.Mood), artW, b))
			fmt.Println()
		}
	}
	if p.URL != "" {
		u := p.URL
		if shown != "" {
			u = strings.Replace(u, "/en/posts/", "/"+shown+"/posts/", 1)
			u = strings.Replace(u, "/de/posts/", "/"+shown+"/posts/", 1)
		}
		printLines([]string{center(dim(s.readMore+" → ")+paint(u, cSky, "4"), w)})
		if o.open {
			openURL(u)
		}
	}
	fmt.Println()
}

func gabTeaser(f *Feed, lang string, w int) {
	for _, m := range f.Models {
		if m.Status == "soon" {
			b := bubble(wrapWords(inline(dict[lang].gabTeaser, "", "3"), w-artW-4), w-artW, moodColor[m.Mood], bold(m.Label, moodColor[m.Mood])+dim(" · soon"), "")
			printLines(sideBySide(garlic(m.Mood), artW, b))
			fmt.Println()
		}
	}
}

func rule(label string, w int) {
	l := " " + label + " "
	left := 3
	right := w - left - vw(l)
	if right < 3 {
		right = 3
	}
	printLines([]string{paint(strings.Repeat("~", left), cMeadow) + paint(l, cPeach, "1") + paint(strings.Repeat("~", right), cMeadow)})
	fmt.Println()
}

func viewDay(f *Feed, day string, o opts, w int, fallback bool) {
	var hits []Post
	for _, p := range f.Posts {
		if p.Date == day {
			hits = append(hits, p)
		}
	}
	rule(fmtDate(day, o.lang), w)
	if len(hits) == 0 {
		b := bubble(wrapWords(inline(dict[o.lang].nothingToday, "", "3"), w-artW-4), w-artW, cPeriwinkle, "", "")
		printLines(sideBySide(garlic("meh"), artW, b))
		fmt.Println()
		if fallback && len(f.Posts) > 0 {
			rule(dict[o.lang].latestInstead, w)
			renderPost(f, f.Posts[0], o, w)
		}
		return
	}
	for i, p := range hits {
		if i > 0 {
			rule("· · ·", w)
		}
		renderPost(f, p, o, w)
	}
	gabTeaser(f, o.lang, w)
}

func viewList(f *Feed, o opts, w int) {
	titleW := w - 2 - 12 - 7 - 7
	rule(fmt.Sprintf("%d posts", len(f.Posts)), w)
	for _, p := range f.Posts {
		t, ok := p.Human[o.lang]
		if !ok {
			for _, x := range p.Human {
				t = x
			}
		}
		title := []rune(t.Title)
		ts := string(title)
		if vw(ts) > titleW {
			ts = string(title[:titleW-1]) + "…"
		}
		var langs []string
		for _, l := range []string{"en", "de"} {
			if _, ok := p.Human[l]; ok {
				langs = append(langs, l)
			}
		}
		ai := 0
		for _, v := range p.Ai {
			if v.Lang == o.lang {
				ai++
			}
		}
		date := fmtDate(p.Date, o.lang)
		row := pad(paint(date, cPeach), 12) + pad(paint(ts, cOat, "1"), titleW+1) + pad(dim(strings.Join(langs, "+")), 7) + paint(fmt.Sprintf("🧄×%d", ai), cMustard)
		printLines([]string{row, "            " + dim(p.Slug)})
	}
	fmt.Println()
	printLines([]string{dim(dict[o.lang].listHint)})
	fmt.Println()
}

func viewHallOfFame(f *Feed, o opts, w int) {
	s := dict[o.lang]
	rule("hall of fame", w)
	printLines([]string{center(paint(s.hofIntro, cOat, "3"), w)})
	fmt.Println()
	titles := map[string]string{}
	for _, p := range f.Posts {
		for l, t := range p.Human {
			titles[p.Slug+"/"+l] = t.Title
		}
	}
	for _, e := range f.HallOfFame {
		m := f.model(e.Model)
		c := moodColor[m.Mood]
		body := wrapWords(inline("“"+e.Quote+"”", c, "1", "3"), w-artW-4)
		if tr, ok := e.Translation[o.lang]; ok && e.Lang != o.lang {
			body = append(body, "")
			body = append(body, wrapWords(append([]word{{dim(s.roughly), vw(s.roughly)}}, inline(tr, cMuted, "3")...), w-artW-4)...)
		}
		title := titles[e.Post+"/"+e.Lang]
		if title == "" {
			title = e.Post
		}
		foot := dim(s.from + ": " + title)
		printLines(sideBySide(garlic(m.Mood), artW, bubble(body, w-artW, c, bold(m.Label, c)+dim(" · "+m.Name), foot)))
		fmt.Println()
	}
}

func openURL(u string) {
	cmd := "xdg-open"
	if runtime.GOOS == "darwin" {
		cmd = "open"
	}
	_ = exec.Command(cmd, u).Start()
}

// find picks a post by slug, a date (YYYY-MM-DD), or words from the slug or title.
func find(f *Feed, q string) (Post, bool) {
	ql := strings.ToLower(q)
	for _, p := range f.Posts {
		if p.Slug == q || p.Date == q {
			return p, true
		}
	}
	for _, p := range f.Posts {
		hay := strings.ToLower(p.Slug)
		for _, t := range p.Human {
			hay += " " + strings.ToLower(t.Title)
		}
		if strings.Contains(hay, ql) {
			return p, true
		}
	}
	return Post{}, false
}

// ── main ───────────────────────────────────────────────────────────────────────

const help = `aioli: daily lab notes from Independence Lab, ruined by tiny AIs.

usage
  aioli                     today's posts (or the latest, if the lab slept)
  aioli --today             only today's posts
  aioli --yesterday         yesterday's posts
  aioli --date 2026-11-12   the posts of one day
  aioli --latest            the newest post
  aioli --list              every post
  aioli <words>             the newest post whose slug or title has these words
  aioli --hof               the hall of fame

options
  --lang en|de              language (default: $AIOLI_LANG, else from $LANG)
  --model <name>            only one tiny AI (135m, 0.5b, qwen, tiny, ...)
  --human                   only the human original
  --ai                      only the AI versions
  --open                    also open the post in the browser
  --no-color, --color       force colours off / on
  --version

  AIOLI_URL overrides the site (default ` + defaultSite + `)
`

func main() {
	args := os.Args[1:]
	o := opts{lang: defaultLang()}
	mode, query, date := "auto", "", ""
	noColor, forceColor := false, false

	need := func(i int, name string) string {
		if i+1 >= len(args) {
			fail("%s needs a value", name)
		}
		return args[i+1]
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if k, v, ok := strings.Cut(a, "="); ok && strings.HasPrefix(a, "--") {
			a = k
			args = append(args[:i+1], append([]string{v}, args[i+1:]...)...)
		}
		switch a {
		case "-h", "--help", "help":
			fmt.Print(help)
			return
		case "-v", "--version", "version":
			fmt.Println("aioli", version)
			return
		case "-t", "--today", "today":
			mode = "today"
		case "-y", "--yesterday", "yesterday":
			mode, date = "date", time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		case "-l", "--latest", "latest":
			mode = "latest"
		case "--list", "list", "ls":
			mode = "list"
		case "--hof", "--hall-of-fame", "hof", "fame":
			mode = "hof"
		case "-d", "--date":
			mode, date = "date", need(i, a)
			i++
		case "--lang":
			o.lang = need(i, a)
			i++
			if o.lang != "en" && o.lang != "de" {
				fail("--lang is en or de")
			}
		case "-m", "--model":
			o.model = need(i, a)
			i++
		case "--human":
			o.humanOnly = true
		case "--ai":
			o.aiOnly = true
		case "-o", "--open":
			o.open = true
		case "--no-color":
			noColor = true
		case "--color":
			forceColor = true
		default:
			if strings.HasPrefix(a, "-") {
				fail("unknown option %s (aioli --help)", a)
			}
			mode = "read"
			if query != "" {
				query += " "
			}
			query += a
		}
	}

	setupColor(forceColor, noColor)
	w := layoutWidth()

	f, stale, err := loadFeed()
	if err != nil {
		fail("couldn't reach the lab: %v", err)
	}

	printLines(banner(w, o.lang))
	if stale {
		when := ""
		if st, err := os.Stat(cachePath()); err == nil {
			when = st.ModTime().Format("2006-01-02 15:04")
		}
		printLines([]string{paint(fmt.Sprintf(dict[o.lang].offline, when), cTomato), ""})
	}

	today := time.Now().Format("2006-01-02")
	switch mode {
	case "auto":
		viewDay(f, today, o, w, true)
	case "today":
		viewDay(f, today, o, w, false)
	case "date":
		if _, err := time.Parse("2006-01-02", date); err != nil {
			fail("dates look like 2026-11-12")
		}
		viewDay(f, date, o, w, false)
	case "latest":
		if len(f.Posts) == 0 {
			fail("no posts yet")
		}
		renderPost(f, f.Posts[0], o, w)
	case "list":
		viewList(f, o, w)
	case "hof":
		viewHallOfFame(f, o, w)
	case "read":
		p, ok := find(f, query)
		if !ok {
			fail(dict[o.lang].noMatch, query)
		}
		renderPost(f, p, o, w)
	}
}

func defaultLang() string {
	if l := os.Getenv("AIOLI_LANG"); l == "en" || l == "de" {
		return l
	}
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			if strings.HasPrefix(v, "de") {
				return "de"
			}
			return "en"
		}
	}
	return "en"
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "🧄 "+format+"\n", a...)
	os.Exit(1)
}
