package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultSite = "https://aioli.darius-852.workers.dev"

type Text struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type AiVersion struct {
	Model   string `json:"model"`
	Lang    string `json:"lang"`
	Persona string `json:"persona"`
	Body    string `json:"body"`
}

type Post struct {
	Slug   string          `json:"slug"`
	Date   string          `json:"date"`
	Author string          `json:"author"`
	URL    string          `json:"url"`
	Human  map[string]Text `json:"human"`
	Ai     []AiVersion     `json:"ai"`
}

type Model struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Mood   string `json:"mood"`
}

type Persona struct {
	ID    string            `json:"id"`
	Label map[string]string `json:"label"`
}

type Fame struct {
	Quote       string            `json:"quote"`
	Translation map[string]string `json:"translation"`
	Model       string            `json:"model"`
	Post        string            `json:"post"`
	Lang        string            `json:"lang"`
}

type Feed struct {
	Version    int       `json:"version"`
	Site       string    `json:"site"`
	Models     []Model   `json:"models"`
	Personas   []Persona `json:"personas"`
	HallOfFame []Fame    `json:"hallOfFame"`
	Posts      []Post    `json:"posts"` // newest first
}

func (f *Feed) model(id string) Model {
	for _, m := range f.Models {
		if m.ID == id {
			return m
		}
	}
	return Model{ID: id, Label: id, Name: id, Mood: "derp"}
}

func (f *Feed) persona(id, lang string) string {
	for _, p := range f.Personas {
		if p.ID == id {
			if l, ok := p.Label[lang]; ok {
				return l
			}
			return p.Label["en"]
		}
	}
	return id
}

func cachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "aioli", "posts.json")
}

// loadFeed fetches /posts.json from the site. When the site can't be reached it falls back to the
// last copy it saved, and says so (stale = true).
func loadFeed() (feed *Feed, stale bool, err error) {
	site := os.Getenv("AIOLI_URL")
	if site == "" {
		site = defaultSite
	}
	raw, fetchErr := fetch(site + "/posts.json")
	cache := cachePath()
	if fetchErr == nil {
		if cache != "" {
			_ = os.MkdirAll(filepath.Dir(cache), 0o755)
			_ = os.WriteFile(cache, raw, 0o644)
		}
	} else {
		if cache == "" {
			return nil, false, fetchErr
		}
		cached, cacheErr := os.ReadFile(cache)
		if cacheErr != nil {
			return nil, false, fetchErr
		}
		raw, stale = cached, true
	}
	feed = &Feed{}
	if err := json.Unmarshal(raw, feed); err != nil {
		return nil, false, fmt.Errorf("the feed came back garbled: %w", err)
	}
	return feed, stale, nil
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "aioli-cli/"+version)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%s answered %s", url, res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, 16<<20))
}
