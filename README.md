# aioli in your terminal

Daily lab notes from Independence Lab, ruined by tiny AIs. The [aioli](https://aioli.darius-852.workers.dev) blog, in a terminal.

```
brew install yukihalf/aioli/aioli                                    # macOS / Linux with Homebrew
curl -fsSL https://aioli.darius-852.workers.dev/install.sh | sh      # anywhere else, no brew needed
go install github.com/YukiHalf/homebrew-aioli/cmd/aioli@latest                  # with Go
```

Then just `aioli`.

```
aioli                     today's posts (or the latest, if the lab slept)
aioli --today             only today's posts
aioli --yesterday         yesterday's posts
aioli --date 2026-11-12   one day
aioli --latest            the newest post
aioli --list              every post
aioli tokenizer           the newest post with "tokenizer" in its slug or title
aioli --hof               the hall of fame

--lang en|de   --model 135m|0.5b|1.1b   --human   --ai   --print   --no-color
```

It opens full screen in the look of the site's ASCII home (shaded garlic, roots into the nav, sticker cards)
and stays until you quit:

```
↑ ↓ j k  space b  g G   scroll (mouse wheel too)
← →  [ ]                newer / older post
t  p  f                 today, all posts, hall of fame
l                       english / deutsch
o                       open the post in the browser
q  Esc                  quit
```

Piped (`aioli | less`) or with `--print` it prints the same page and exits.

It reads `/posts.json` from the site (built with the site, so it is as fresh as the pages) and keeps the last copy,
so it still works offline. `AIOLI_URL` points it at another copy of the site, e.g. a local `astro preview`.

Without brew: `go install github.com/YukiHalf/homebrew-aioli/cmd/aioli@latest` (Go 1.22+), or `go build -o aioli ./cmd/aioli` here.

## Releasing

```
git tag v0.2.0 && git push origin v0.2.0
```

The release workflow builds macOS + Linux binaries (arm64 and amd64), publishes them as a GitHub release with
checksums, and rewrites `Formula/aioli.rb` to point at them. `brew upgrade aioli` and the install script pick it up.
