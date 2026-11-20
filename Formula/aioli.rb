class Aioli < Formula
  desc "Daily lab notes from Independence Lab, ruined by tiny AIs, in your terminal"
  homepage "https://aioli.darius-852.workers.dev"
  url "https://github.com/YukiHalf/homebrew-aioli/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_AFTER_TAGGING"
  head "https://github.com/YukiHalf/homebrew-aioli.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/aioli --version")
  end
end
