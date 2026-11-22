class Aioli < Formula
  desc "Daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal"
  homepage "https://aioli.darius-852.workers.dev"
  version "0.1.0"
  head "https://github.com/YukiHalf/homebrew-aioli.git", branch: "main"

  base = "https://github.com/YukiHalf/homebrew-aioli/releases/download/v0.1.0/aioli_0.1.0"
  on_macos do
    on_arm do
      url "#{base}_darwin_arm64.tar.gz"
      sha256 "ff6fc65d1739ed75457e27413b31b357a3b166dc8ead99f9422c72f98422c77b"
    end
    on_intel do
      url "#{base}_darwin_amd64.tar.gz"
      sha256 "7e5f25cc337aa7a4c50120c6c61baa1a7169510f49b81e29ea3ef1b1cc637f2d"
    end
  end
  on_linux do
    on_arm do
      url "#{base}_linux_arm64.tar.gz"
      sha256 "39c851b797ad6539e7b45e1d5ac763cf3194319231a0005537f9ffedca92711e"
    end
    on_intel do
      url "#{base}_linux_amd64.tar.gz"
      sha256 "93fc8e0a73ddacd87a1883acfe8cf36a4d0d7f031e5ad745bda0a8e1f029aba0"
    end
  end

  head do
    depends_on "go" => :build
  end

  def install
    if build.head?
      system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=HEAD"), "./cmd/aioli"
    else
      bin.install "aioli"
    end
  end

  test do
    assert_match "aioli", shell_output("#{bin}/aioli --version")
  end
end
