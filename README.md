# aioli in your terminal

Daily lab notes from Independence Lab, ruined by tiny AIs. The [aioli](https://aioli.darius-852.workers.dev) blog, in a terminal.

```
brew install yukihalf/aioli/aioli
aioli --today
```

```
aioli                     today's posts (or the latest, if the lab slept)
aioli --today             only today's posts
aioli --yesterday         yesterday's posts
aioli --date 2026-11-12   one day
aioli --latest            the newest post
aioli --list              every post
aioli tokenizer           the newest post with "tokenizer" in its slug or title
aioli --hof               the hall of fame

--lang en|de   --model 135m|0.5b|1.1b   --human   --ai   --open   --no-color
```

It reads `/posts.json` from the site (built with the site, so it is as fresh as the pages) and keeps the last copy,
so it still works offline. `AIOLI_URL` points it at another copy of the site, e.g. a local `astro preview`.

Without brew: `go install github.com/YukiHalf/homebrew-aioli@latest` (Go 1.22+), or `go build -o aioli .` here.

## Releasing

```
git tag v0.2.0 && git push --tags
curl -sL https://github.com/YukiHalf/homebrew-aioli/archive/refs/tags/v0.2.0.tar.gz | shasum -a 256
```

Put the new tag and checksum into `Formula/aioli.rb`, commit, push. `brew upgrade aioli` picks it up.
