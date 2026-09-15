class Pickbin < Formula
  desc "Pick the right binary from any GitHub release for your machine"
  homepage "https://github.com/mmmmaharshi/pickbin"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    else
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-linux-arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-linux-amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "pickbin-darwin-#{Hardware::CPU.arm? ? "arm64" : "amd64"}" => "pickbin" if OS.mac?
    bin.install "pickbin-linux-#{Hardware::CPU.arm? ? "arm64" : "amd64"}" => "pickbin" if OS.linux?
  end

  test do
    assert_match "Usage", shell_output("#{bin}/pickbin 2>&1", 1)
  end
end
