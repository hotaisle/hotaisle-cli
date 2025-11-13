# This is a template file used to generate the brew formula.
# https://github.com/hotaisle/homebrew-tap/blob/main/hotaisle.rb
class HotaisleCli < Formula
  desc "Hotaisle CLI tool"
  homepage "https://github.com/hotaisle/hotaisle-cli"
  version "VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/hotaisle/hotaisle-cli/releases/download/VERSION/hotaisle-cli-VERSION-darwin-arm64.tar.gz"
      sha256 "ARM64_SHA256"
    else
      url "https://github.com/hotaisle/hotaisle-cli/releases/download/VERSION/hotaisle-cli-VERSION-darwin-amd64.tar.gz"
      sha256 "AMD64_SHA256"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "hotaisle-cli-darwin-arm64" => "hotaisle"
    else
      bin.install "hotaisle-cli-darwin-amd64" => "hotaisle"
    end
  end

  test do
    system "#{bin}/hotaisle"
    assert_match version.to_s, shell_output("#{bin}/hotaisle --version")
  end
end
