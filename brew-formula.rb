class HotaisleCli < Formula
  desc "Hotaisle CLI tool"
  homepage "https://github.com/hotaisle/hotaisle-cli"
  version "VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/hotaisle/hotaisle-cli/releases/download/vVERSION/hotaisle-cli-darwin-arm64.tar.gz"
      sha256 "ARM64_SHA256"
    else
      url "https://github.com/hotaisle/hotaisle-cli/releases/download/vVERSION/hotaisle-cli-darwin-amd64.tar.gz"
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
    system "/usr/local/bin/hotaisle", "--version"
  end
end
