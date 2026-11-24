class Aioli < Formula
  desc "Daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal"
  homepage "https://aioli.darius-852.workers.dev"
  license "MIT"
  head "https://github.com/YukiHalf/homebrew-aioli.git", branch: "main"

  base = "https://github.com/YukiHalf/homebrew-aioli/releases/download/v0.1.1/aioli_0.1.1"
  on_macos do
    on_arm do
      url "#{base}_darwin_arm64.tar.gz"
      sha256 "7b1208fd81c91dd17f7df79b36b32765e4cf51295ead9d0abb6755ba2381041a"
    end
    on_intel do
      url "#{base}_darwin_amd64.tar.gz"
      sha256 "de408d2fea8d2b55a0075812df03fee6fd5f2b86febce635c4f1ecbed0687955"
    end
  end
  on_linux do
    on_arm do
      url "#{base}_linux_arm64.tar.gz"
      sha256 "ec0489f2864630310e8ae52305dcb3bc4045724f7b1534967e416ad5d1c9213f"
    end
    on_intel do
      url "#{base}_linux_amd64.tar.gz"
      sha256 "d4142bd97cf1aa30d61a4ae323f978c2bcd9ec2102e1adb85d8085973cdf6e3d"
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
