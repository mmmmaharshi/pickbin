class Pickbin < Formula
  desc "Pick the right binary from any GitHub release for your machine"
  homepage "https://github.com/mmmmaharshi/pickbin"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-darwin-arm64"
      sha256 "6459fa9f9bd6edd69fb1ffd70812a3017d7c907366ce3a4bc2bab09b4f974ecd"
    else
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-darwin-amd64"
      sha256 "796a67940669a5b0ef428b035f0787dac6cc5ef46161ea3ad63c0120e9f681bb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-linux-arm64"
      sha256 "0e9c4f2d47374e0038510e87fa4f33b92e6c908c12a60e309d426017f1a69747"
    else
      url "https://github.com/mmmmaharshi/pickbin/releases/download/v0.1.0/pickbin-linux-amd64"
      sha256 "cff4e1ecc711bec67a3afba05597a59ade03bc6654704daabeb397e040630283"
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
