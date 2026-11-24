class Aioli < Formula
  desc "Daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal"
  homepage "https://aioli.darius-852.workers.dev"
  license "MIT"
  head "https://github.com/YukiHalf/homebrew-aioli.git", branch: "main"

  base = "https://github.com/YukiHalf/homebrew-aioli/releases/download/v0.1.2/aioli_0.1.2"
  on_macos do
    on_arm do
      url "#{base}_darwin_arm64.tar.gz"
      sha256 "8de75807577a143a3e83b052e6522cd56bae288f61688521d8c168b7bdbb8b85"
    end
    on_intel do
      url "#{base}_darwin_amd64.tar.gz"
      sha256 "1dabac7bd1bde12e83433900d61b0d9e7bcb19c02c0b0c3dcfe7b6944fcf01f3"
    end
  end
  on_linux do
    on_arm do
      url "#{base}_linux_arm64.tar.gz"
      sha256 "e3dfeae7a37fa0e8dfd3c32fb514010c014760f082436d468977dabc8262a1c0"
    end
    on_intel do
      url "#{base}_linux_amd64.tar.gz"
      sha256 "efbecd025d036865d47efeee2103db5569da38a0489e2aac9b5b00b14beeab78"
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
