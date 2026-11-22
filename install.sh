#!/bin/sh
# aioli installer: downloads the newest release binary for this machine into ~/.local/bin (or $AIOLI_BIN).
#   curl -fsSL https://aioli.darius-852.workers.dev/install.sh | sh
set -eu
repo=YukiHalf/homebrew-aioli
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in darwin|linux) ;; *) echo "aioli: no build for $os (try: go install github.com/$repo@latest)" >&2; exit 1 ;; esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "aioli: no build for $(uname -m)" >&2; exit 1 ;;
esac
tag=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
[ -n "$tag" ] || { echo "aioli: couldn't find the latest release" >&2; exit 1; }
v=${tag#v}
name="aioli_${v}_${os}_${arch}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "🧄 downloading aioli $v ($os/$arch)"
curl -fsSL -o "$tmp/$name.tar.gz" "https://github.com/$repo/releases/download/$tag/$name.tar.gz"
curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$repo/releases/download/$tag/checksums.txt"
want=$(grep "$name.tar.gz" "$tmp/checksums.txt" | cut -d' ' -f1)
if command -v sha256sum >/dev/null; then got=$(sha256sum "$tmp/$name.tar.gz" | cut -d' ' -f1); else got=$(shasum -a 256 "$tmp/$name.tar.gz" | cut -d' ' -f1); fi
[ "$want" = "$got" ] || { echo "aioli: checksum mismatch, not installing" >&2; exit 1; }
tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
dir=${AIOLI_BIN:-$HOME/.local/bin}
mkdir -p "$dir"
install -m 755 "$tmp/$name/aioli" "$dir/aioli"
echo "🧄 installed $dir/aioli"
case ":$PATH:" in *":$dir:"*) ;; *) echo "   add it to your PATH:  export PATH=\"$dir:\$PATH\"" ;; esac
echo "   run: aioli"
