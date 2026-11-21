// aioli: daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var version = "dev"

// ── words, in both languages ───────────────────────────────────────────────────

type words struct {
	tagline, human, writtenAs, onlyIn, aiWarning, gabTeaser, nothingToday, latestInstead,
	noMatch, alsoIn, ruinedBy, hofIntro, from, roughly, offline, readMore, listHint string
	navToday, navPosts, navFame, heroNote, by, aiVersion, aiVersions, keys string
	hero                                                                   [3]string
	months                                                                 []string
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
		listHint:      "← → read them one by one · or: aioli <a word from the title>",
		navToday:      "today",
		navPosts:      "posts",
		navFame:       "hall of fame",
		heroNote:      "scroll v",
		by:            "by",
		aiVersion:     "ai version",
		aiVersions:    "ai versions",
		keys:          "←→ posts · t today · p list · f fame · l DE · q quit",
		hero:          [3]string{"daily lab notes, ", "ruined", " by tiny AIs"},
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
		listHint:      "← → einer nach dem anderen · oder: aioli <ein Wort aus dem Titel>",
		navToday:      "heute",
		navPosts:      "posts",
		navFame:       "hall of fame",
		heroNote:      "scrollen v",
		by:            "von",
		aiVersion:     "KI-Version",
		aiVersions:    "KI-Versionen",
		keys:          "←→ Posts · t heute · p Liste · f Fame · l EN · q raus",
		hero:          [3]string{"tägliche Lab-Notizen, ", "ruiniert", " von winzigen KIs"},
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

// ── state ──────────────────────────────────────────────────────────────────────

type opts struct {
	model     string
	humanOnly bool
	aiOnly    bool
}

// state is what the screen shows; the nav keys change it and the page is rendered again.
type state struct {
	f        *Feed
	stale    bool
	lang     string
	mode     string // today | day | post | list | hof
	date     string
	fallback bool // today with no posts: show the latest instead
	post     int  // index into f.Posts (post mode, and where ← → start from)
	isToday  bool
	o        opts

	dumber    int   // easter egg: how many times d was pressed on this post
	seed      int64 // its dice
	laugh     bool  // easter egg: the garlic is laughing (Konami code)
	konamiPos int
}

func (st *state) command() string {
	c := "aioli"
	switch st.mode {
	case "today":
		c += " --today"
	case "day":
		c += " --date " + st.date
	case "post":
		p := st.f.Posts[st.post]
		c += " " + strings.TrimLeft(strings.TrimPrefix(p.Slug, p.Date), "-")
	case "list":
		c += " --list"
	case "hof":
		c += " --hof"
	}
	if st.lang == "de" {
		c += " --lang de"
	}
	return c
}

func (st *state) title() string {
	switch st.mode {
	case "today", "day":
		return fmtDate(st.date, st.lang)
	case "post":
		return fmtDate(st.f.Posts[st.post].Date, st.lang)
	case "list":
		return dict[st.lang].navPosts
	case "hof":
		return dict[st.lang].navFame
	}
	return ""
}

// everything a view prints lands in buf
var buf []string

func printLines(ls []string) { buf = append(buf, ls...) }
func nl()                    { buf = append(buf, "") }

func (st *state) render(w, rows int) []string {
	buf = nil
	printLines(hero(st, w, rows))
	if st.stale {
		when := ""
		if fi, err := os.Stat(cachePath()); err == nil {
			when = fi.ModTime().Format("2006-01-02 15:04")
		}
		printLines([]string{"  " + paint(fmt.Sprintf(dict[st.lang].offline, when), cTomato), ""})
	}
	switch st.mode {
	case "today", "day":
		viewDay(st, w)
	case "post":
		renderPost(st, st.post, w)
		if st.dumber > 0 {
			renderDumber(st, w)
		}
	case "list":
		viewList(st, w)
	case "hof":
		viewHallOfFame(st, w)
	}
	printLines(footer(st.lang, w))
	return buf
}

// ── views ──────────────────────────────────────────────────────────────────────

func section(label string) {
	printLines([]string{"  " + paint("[ ", cMustard) + paint(label, cMustard, "1") + paint(" ]", cMustard), ""})
}

func postID(st *state, i int) string {
	return "#" + fmt.Sprintf("%03d", len(st.f.Posts)-i)
}

func aiCount(s words, n int) string {
	if n == 1 {
		return "1 " + s.aiVersion
	}
	return strconv.Itoa(n) + " " + s.aiVersions
}

// shownLang: the reader's language if the post has it, else the one it was written in.
func shownLang(p Post, lang string) (string, Text) {
	if t, ok := p.Human[lang]; ok {
		return lang, t
	}
	for _, l := range []string{"en", "de"} {
		if t, ok := p.Human[l]; ok {
			return l, t
		}
	}
	return lang, Text{}
}

func versionsFor(st *state, p Post, lang string) []AiVersion {
	var out []AiVersion
	for _, v := range p.Ai {
		if v.Lang != lang {
			continue
		}
		if st.o.model != "" {
			m := st.f.model(v.Model)
			if !strings.Contains(strings.ToLower(m.ID+" "+m.Label+" "+m.Name), strings.ToLower(st.o.model)) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

// renderPost: the human original as a big card, then each tiny AI's version as its own card.
func renderPost(st *state, idx int, w int) {
	s := dict[st.lang]
	p := st.f.Posts[idx]
	shown, text := shownLang(p, st.lang)
	versions := versionsFor(st, p, shown)
	color := cardPalette[idx%len(cardPalette)]
	cw := w - 3

	if !st.o.aiOnly {
		var stickers []string
		for _, v := range versions {
			face, c := faceOf(v.Model)
			stickers = append(stickers, chip(face, cSpace, c))
		}
		lines := []string{paint(postID(st, idx), color)}
		lines = append(lines, wrapWords(inline(text.Title, color, "1"), cw-4)...)
		meta := fmtDate(p.Date, st.lang) + " · " + s.by + " " + p.Author
		if len(versions) > 0 {
			meta += " · " + aiCount(s, len(versions))
		}
		lines = append(lines, paint(meta, cCream), "")
		lines = append(lines, renderMarkdown(text.Body, cw-4, cCream)...)
		bottom := ""
		if shown != st.lang {
			bottom = paint(fmt.Sprintf(s.onlyIn, langName(shown, st.lang)), cTangerine)
		} else if len(p.Human) > 1 {
			bottom = dim(s.alsoIn + " " + langName(otherLang(shown), st.lang) + ": l")
		}
		printLines(indent(card(lines, cw, color, stickers, bottom), 1))
		nl()
	}

	if !st.o.humanOnly && len(versions) > 0 {
		if !st.o.aiOnly {
			printLines([]string{
				"     " + paint("v ", cTangerine) + paint(s.ruinedBy, cLavender, "3"),
				"     " + paint("! "+s.aiWarning, cTangerine, "3"),
				"",
			})
		}
		room := 8
		for i, v := range versions {
			m := st.f.model(v.Model)
			face, c := faceOf(v.Model)
			lines := []string{bold(m.Label, c) + dim(" · "+m.Name), ""}
			lines = append(lines, renderMarkdown(v.Body, cw-room-4, cCream)...)
			b := card(lines, cw-room, c, []string{chip(face, cSpace, c)}, dim(s.writtenAs+" "+st.f.persona(v.Persona, st.lang)))
			printLines(indent(b, 1+staggerOf(i+1, room)))
			nl()
		}
	}
	if p.URL != "" {
		printLines([]string{"  " + dim(s.readMore+" → ") + paint(postURL(p, shown), cSky, "4")})
	}
	nl()
}

func otherLang(l string) string {
	if l == "de" {
		return "en"
	}
	return "de"
}

func postURL(p Post, lang string) string {
	u := strings.Replace(p.URL, "/en/posts/", "/"+lang+"/posts/", 1)
	return strings.Replace(u, "/de/posts/", "/"+lang+"/posts/", 1)
}

func viewDay(st *state, w int) {
	var hits []int
	for i, p := range st.f.Posts {
		if p.Date == st.date {
			hits = append(hits, i)
		}
	}
	section(fmtDate(st.date, st.lang))
	if len(hits) == 0 {
		face, c := faceOf("tinyllama-1.1b")
		b := card(wrapWords(inline(dict[st.lang].nothingToday, cCream, "3"), w-12), w-8, c, []string{chip(face, cSpace, c)}, "")
		printLines(indent(b, 1))
		nl()
		if st.fallback && len(st.f.Posts) > 0 {
			section(dict[st.lang].latestInstead)
			renderPost(st, 0, w)
		}
		return
	}
	for _, i := range hits {
		renderPost(st, i, w)
	}
}

func viewList(st *state, w int) {
	s := dict[st.lang]
	section(fmt.Sprintf("%d posts", len(st.f.Posts)))
	cw := w - 3 - 12
	for i, p := range st.f.Posts {
		shown, t := shownLang(p, st.lang)
		versions := versionsFor(st, p, shown)
		color := cardPalette[i%len(cardPalette)]
		var stickers []string
		for _, v := range versions {
			face, c := faceOf(v.Model)
			stickers = append(stickers, chip(face, cSpace, c))
		}
		title := wrapWords(inline(t.Title, color, "1"), cw-4)
		if shown != st.lang {
			title = append(title, paint(fmt.Sprintf(s.onlyIn, langName(shown, st.lang)), cLavender, "3"))
		}
		lines := append([]string{paint(postID(st, i), color)}, title...)
		meta := fmtDate(p.Date, st.lang) + " · " + s.by + " " + p.Author
		if len(versions) > 0 {
			meta += " · " + aiCount(s, len(versions))
		}
		lines = append(lines, paint(meta, cCream))
		printLines(indent(card(lines, cw, color, stickers, ""), 1+staggerOf(i, 12)))
		nl()
	}
	printLines([]string{"  " + dim(s.listHint)})
}

func viewHallOfFame(st *state, w int) {
	s := dict[st.lang]
	section(s.navFame)
	printLines([]string{"  " + paint(s.hofIntro, cCream, "3"), ""})
	titles := map[string]string{}
	for _, p := range st.f.Posts {
		for l, t := range p.Human {
			titles[p.Slug+"/"+l] = t.Title
		}
	}
	cw := w - 3 - 8
	for i, e := range st.f.HallOfFame {
		m := st.f.model(e.Model)
		face, c := faceOf(e.Model)
		lines := []string{bold(m.Label, c) + dim(" · "+m.Name), ""}
		lines = append(lines, wrapWords(inline("“"+e.Quote+"”", c, "1", "3"), cw-4)...)
		if tr, ok := e.Translation[st.lang]; ok && e.Lang != st.lang {
			lines = append(lines, "")
			lines = append(lines, wrapWords(append([]word{{dim(s.roughly), vw(s.roughly)}}, inline(tr, cCream, "3")...), cw-4)...)
		}
		title := titles[e.Post+"/"+e.Lang]
		if title == "" {
			title = e.Post
		}
		printLines(indent(card(lines, cw, c, []string{chip(face, cSpace, c)}, dim(s.from+": "+title)), 1+staggerOf(i, 8)))
		nl()
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
func find(f *Feed, q string) (int, bool) {
	ql := strings.ToLower(q)
	for i, p := range f.Posts {
		if p.Slug == q || p.Date == q {
			return i, true
		}
	}
	for i, p := range f.Posts {
		hay := strings.ToLower(p.Slug)
		for _, t := range p.Human {
			hay += " " + strings.ToLower(t.Title)
		}
		if strings.Contains(hay, ql) {
			return i, true
		}
	}
	return 0, false
}

// currentPost: the post on screen (the first one, on a day with several).
func (st *state) currentPost() int {
	if st.mode == "today" || st.mode == "day" {
		for i, p := range st.f.Posts {
			if p.Date == st.date {
				return i
			}
		}
		return 0
	}
	if st.mode == "post" {
		return st.post
	}
	return 0
}

// onKey: the nav. Returns whether the page has to be rendered again from the top.
func (st *state) onKey(k int) bool {
	n := len(st.f.Posts)
	st.laugh = false
	if k != 'd' {
		st.dumber = 0
	}
	switch k {
	case 'd':
		if n == 0 {
			return false
		}
		st.post = st.currentPost()
		st.mode = "post"
		st.dumber++
		st.seed = newSeed()
	case kRight, ']', '>': // older
		if n == 0 {
			return false
		}
		if st.mode == "post" {
			if st.post+1 >= n {
				return false
			}
			st.post++
		}
		st.mode = "post"
	case kLeft, '[', '<': // newer
		if n == 0 {
			return false
		}
		if st.mode == "post" {
			if st.post == 0 {
				return false
			}
			st.post--
		}
		st.mode = "post"
	case 't':
		st.mode, st.date, st.isToday, st.fallback = "today", time.Now().Format("2006-01-02"), true, true
	case 'p':
		st.mode = "list"
	case 'f':
		st.mode = "hof"
	case 'l', 'L':
		st.lang = otherLang(st.lang)
	case 'o':
		if n > 0 {
			p := st.f.Posts[st.post]
			l, _ := shownLang(p, st.lang)
			openURL(postURL(p, l))
		}
		return false
	default:
		return false
	}
	return true
}

func (st *state) status() string {
	return dict[st.lang].keys
}

func (st *state) statusShort() string {
	if st.lang == "de" {
		return "←→ t p f l o · q raus"
	}
	return "←→ t p f l o · q quit"
}

// ── main ───────────────────────────────────────────────────────────────────────

const help = `aioli: daily lab notes from Independence Lab, ruined by tiny AIs.

usage   (opens full screen; q quits)
  aioli                     today's posts (or the latest, if the lab slept)
  aioli --today             only today's posts
  aioli --yesterday         yesterday's posts
  aioli --date 2026-11-12   the posts of one day
  aioli --latest            the newest post
  aioli --list              every post
  aioli <words>             the newest post whose slug or title has these words
  aioli --hof               the hall of fame

keys
  ↑ ↓ j k  space b  g G     scroll, page, top / end (mouse wheel too)
  ← →  [ ]                  newer / older post
  t  p  f                   today, all posts, hall of fame
  l                         english / deutsch
  o                         open the post in the browser
  q  Esc                    quit

options
  --lang en|de              language (default: $AIOLI_LANG, else from $LANG)
  --model <name>            only one tiny AI (135m, 0.5b, qwen, tiny, ...)
  --human                   only the human original
  --ai                      only the AI versions
  --print                   print and exit instead of the full-screen view
                            (automatic when piped)
  --no-color, --color       force colours off / on
  --version

  AIOLI_URL overrides the site (default ` + defaultSite + `)
`

func main() {
	args := os.Args[1:]
	st := &state{lang: defaultLang(), mode: "today", fallback: true}
	query := ""
	noColor, forceColor, noPager := false, false, false

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
		case "sudo":
			sudo(st.lang)
			return
		case "-h", "--help", "help":
			fmt.Print(help)
			return
		case "-v", "--version", "version":
			fmt.Println("aioli", version)
			return
		case "-t", "--today", "today":
			st.mode, st.fallback = "today", false
		case "-y", "--yesterday", "yesterday":
			st.mode, st.date = "day", time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		case "-l", "--latest", "latest":
			st.mode = "post"
		case "--list", "list", "ls":
			st.mode = "list"
		case "--hof", "--hall-of-fame", "hof", "fame":
			st.mode = "hof"
		case "-d", "--date":
			st.mode, st.date = "day", need(i, a)
			i++
			if _, err := time.Parse("2006-01-02", st.date); err != nil {
				fail("dates look like 2026-11-12")
			}
		case "--lang":
			st.lang = need(i, a)
			i++
			if st.lang != "en" && st.lang != "de" {
				fail("--lang is en or de")
			}
		case "-m", "--model":
			st.o.model = need(i, a)
			i++
		case "--dumber": // easter egg, not in --help
			n, err := strconv.Atoi(need(i, a))
			i++
			if err != nil || n < 1 {
				fail("--dumber wants a number from 1")
			}
			st.dumber, st.seed = n, newSeed()
			if st.mode == "today" {
				st.mode = "post"
			}
		case "--human":
			st.o.humanOnly = true
		case "--ai":
			st.o.aiOnly = true
		case "-p", "--print", "--no-pager":
			noPager = true
		case "--no-color":
			noColor = true
		case "--color":
			forceColor = true
		default:
			if strings.HasPrefix(a, "-") {
				fail("unknown option %s (aioli --help)", a)
			}
			st.mode = "read"
			if query != "" {
				query += " "
			}
			query += a
		}
	}

	setupColor(forceColor, noColor)
	f, stale, err := loadFeed()
	if err != nil {
		fail("couldn't reach the lab: %v", err)
	}
	st.f, st.stale = f, stale

	today := time.Now().Format("2006-01-02")
	switch st.mode {
	case "today":
		st.date, st.isToday = today, true
	case "day":
		st.isToday = st.date == today
	case "read":
		i, ok := find(f, query)
		if !ok {
			fail(dict[st.lang].noMatch, query)
		}
		st.mode, st.post = "post", i
	}
	if st.mode == "post" && len(f.Posts) == 0 {
		fail("no posts yet")
	}

	if isTTY() && !noPager {
		page(st)
		return
	}
	rows := 40
	if r, _, ok := termRowsCols(); ok {
		rows = r
	}
	for _, l := range st.render(layoutWidth(), rows) {
		fmt.Println(strings.TrimRight(l, " "))
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
	fmt.Fprintf(os.Stderr, "aioli: "+format+"\n", a...)
	os.Exit(1)
}
