# Package Manager Installation

This document outlines the package manager submission process for WTF.

## 🍺 Homebrew (macOS/Linux)

### Creating the Formula

1. Fork the [homebrew-core](https://github.com/Homebrew/homebrew-core) repository
2. Create a formula file at `Formula/wtf.rb`:

```ruby
class Wtf < Formula
  desc "CLI tool to discover shell commands using natural language"
  homepage "https://github.com/Vedant9500/WTF"
  url "https://github.com/Vedant9500/WTF/archive/v1.1.0.tar.gz"
  sha256 "YOUR_SHA256_HERE"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/wtf"
  end

  test do
    assert_match "wtf version", shell_output("#{bin}/wtf --version")
  end
end
```

3. Submit a pull request to homebrew-core

### Installation Command
```bash
brew install wtf
```

## 🍫 Chocolatey (Windows)

### Creating the Package

1. Create a `tools/chocolateyinstall.ps1`:

```powershell
$ErrorActionPreference = 'Stop'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = 'https://github.com/Vedant9500/WTF/releases/download/v1.1.0/wtf-windows-amd64.zip'

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  softwareName  = 'wtf*'
  checksum64    = 'YOUR_CHECKSUM_HERE'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
```

2. Create a `wtf.nuspec`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>wtf</id>
    <version>1.1.0</version>
    <packageSourceUrl>https://github.com/Vedant9500/WTF</packageSourceUrl>
    <owners>Vedant9500</owners>
    <title>WTF (What's The Function)</title>
    <authors>Vedant9500</authors>
    <projectUrl>https://github.com/Vedant9500/WTF</projectUrl>
    <iconUrl>https://cdn.rawgit.com/Vedant9500/WTF/main/icon.png</iconUrl>
    <copyright>2025 Vedant9500</copyright>
    <licenseUrl>https://github.com/Vedant9500/WTF/blob/main/LICENSE</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
    <projectSourceUrl>https://github.com/Vedant9500/WTF</projectSourceUrl>
    <docsUrl>https://github.com/Vedant9500/WTF/blob/main/README.md</docsUrl>
    <bugTrackerUrl>https://github.com/Vedant9500/WTF/issues</bugTrackerUrl>
    <tags>cli shell command discovery natural-language</tags>
    <summary>CLI tool to discover shell commands using natural language</summary>
    <description>WTF is a powerful CLI tool that helps developers discover shell commands using advanced natural language processing. When you can't remember a command, you think "What's The Function I need?" - that's WTF!</description>
  </metadata>
  <files>
    <file src="tools\**" target="tools" />
  </files>
</package>
```

### Installation Command
```powershell
choco install wtf
```

## 📦 Snap (Linux)

### Creating the Snap Package

1. Create a `snap/snapcraft.yaml`:

```yaml
name: wtf
base: core20
version: '1.1.0'
summary: CLI tool to discover shell commands using natural language
description: |
  WTF is a powerful CLI tool that helps developers discover shell commands 
  using advanced natural language processing. When you can't remember a command, 
  you think "What's The Function I need?" - that's WTF!

grade: stable
confinement: strict

parts:
  wtf:
    plugin: go
    source: .
    source-type: git
    build-snaps: [go]
    go-importpath: github.com/Vedant9500/WTF

apps:
  wtf:
    command: bin/wtf
    plugs: [home, network]
```

### Installation Command
```bash
sudo snap install wtf
```

## 🐳 Docker

### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wtf ./cmd/wtf

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/wtf .
COPY --from=builder /app/assets ./assets
CMD ["./wtf"]
```

### Docker Hub Publication

```bash
# Build and push
docker build -t vedant9500/wtf:v1.1.0 .
docker tag vedant9500/wtf:v1.1.0 vedant9500/wtf:latest
docker push vedant9500/wtf:v1.1.0
docker push vedant9500/wtf:latest
```

### Installation Command
```bash
docker run -it --rm vedant9500/wtf:latest "compress files"
```

## 📱 Other Package Managers

### AUR (Arch Linux)

Create a PKGBUILD file for the Arch User Repository.

### APT Repository (Ubuntu/Debian)

Set up a custom APT repository for .deb packages.

### RPM Repository (Red Hat/CentOS/Fedora)

Create RPM packages for Red Hat-based distributions.

## 🚀 Automation

The release workflow automatically creates all necessary assets. To submit to package managers:

1. **Wait for release**: Let GitHub Actions build all binaries
2. **Download checksums**: Get SHA256 checksums from the release
3. **Update package configs**: Use the checksums in package manager configs
4. **Submit packages**: Create pull requests to respective package repositories

## 📋 Checklist for Package Manager Submission

- [ ] Homebrew formula created and tested
- [ ] Chocolatey package created and tested  
- [ ] Snap package created and tested
- [ ] Docker image built and pushed
- [ ] AUR package created
- [ ] Documentation updated with installation instructions
