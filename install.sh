#!/bin/sh
# aioli installer: no root needed. Downloads the newest release binary for this machine into ~/.local/bin
# (or $AIOLI_BIN) and puts that folder on your PATH (skip with AIOLI_NO_MODIFY_PATH=1).
#   curl -fsSL https://aioli.darius-852.workers.dev/install.sh | sh
#   wget -qO-  https://aioli.darius-852.workers.dev/install.sh | sh
set -eu
repo=YukiHalf/homebrew-aioli
die() { echo "aioli: $*" >&2; exit 1; }

if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL -o "$2" "$1"; }
  # the release page redirects to /releases/tag/vX.Y.Z; no GitHub API, so shared IPs don't hit its rate limit
  latest() { curl -fsSLI -o /dev/null -w '%{url_effective}' "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO "$2" "$1"; }
  latest() { wget -S --spider "$1" 2>&1 | sed -n 's/^ *[Ll]ocation: *\([^ ]*\).*/\1/p' | tail -1; }
else
  die "needs curl or wget"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in darwin|linux) ;; *) die "no build for $os (with Go: go install github.com/$repo/cmd/aioli@latest)" ;; esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) die "no build for $(uname -m) (with Go: go install github.com/$repo/cmd/aioli@latest)" ;;
esac

tag=${AIOLI_VERSION:-$(latest "https://github.com/$repo/releases/latest" | tr -d '\r' | sed 's#.*/tag/##')}
case "$tag" in v[0-9]*) ;; *) die "couldn't find the latest release" ;; esac
v=${tag#v}
name="aioli_${v}_${os}_${arch}"
tmp=$(mktemp -d 2>/dev/null || mktemp -d -t aioli)
trap 'rm -rf "$tmp"' EXIT
echo "🧄 downloading aioli $v ($os/$arch)"
fetch "https://github.com/$repo/releases/download/$tag/$name.tar.gz" "$tmp/$name.tar.gz" || die "download failed"
fetch "https://github.com/$repo/releases/download/$tag/checksums.txt" "$tmp/checksums.txt" || die "download failed"
want=$(grep "$name.tar.gz" "$tmp/checksums.txt" | cut -d' ' -f1)
if command -v sha256sum >/dev/null 2>&1; then got=$(sha256sum "$tmp/$name.tar.gz" | cut -d' ' -f1)
elif command -v shasum >/dev/null 2>&1; then got=$(shasum -a 256 "$tmp/$name.tar.gz" | cut -d' ' -f1)
else got=$(openssl dgst -sha256 "$tmp/$name.tar.gz" | sed 's/.*= *//'); fi
[ -n "$want" ] && [ "$want" = "$got" ] || die "checksum mismatch, not installing"
tar -xzf "$tmp/$name.tar.gz" -C "$tmp"

dir=${AIOLI_BIN:-$HOME/.local/bin}
mkdir -p "$dir" || die "can't create $dir (set AIOLI_BIN to a folder you own)"
cp "$tmp/$name/aioli" "$dir/aioli.new" && chmod 755 "$dir/aioli.new" && mv -f "$dir/aioli.new" "$dir/aioli"
echo "🧄 installed $dir/aioli"

case ":$PATH:" in
  *":$dir:"*) echo "   run: aioli" ;;
  *)
    line="export PATH=\"$dir:\$PATH\""
    if [ "${AIOLI_NO_MODIFY_PATH:-}" = 1 ]; then
      echo "   add it to your PATH:  $line"
    else
      case "$(basename "${SHELL:-sh}")" in
        zsh)  rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
        bash) rc="$HOME/.bashrc" ;;
        fish) rc="" ; mkdir -p "$HOME/.config/fish/conf.d" && echo "fish_add_path $dir" > "$HOME/.config/fish/conf.d/aioli.fish" ;;
        *)    rc="$HOME/.profile" ;;
      esac
      if [ -n "$rc" ] && ! grep -qs "# added by aioli" "$rc"; then
        printf '\n%s  # added by aioli\n' "$line" >> "$rc"
      fi
      echo "   added $dir to your PATH; open a new terminal, or run now:  $line"
    fi
    echo "   then: aioli" ;;
esac
