# Maintainer: Vedant9500 <your-email@example.com>
pkgname=wtf-cli
pkgver=1.1.0
pkgrel=1
pkgdesc="WTF (What's The Function) - CLI command discovery tool with advanced NLP"
arch=('x86_64' 'aarch64')
url="https://github.com/Vedant9500/WTF"
license=('MIT')
depends=()
makedepends=('go' 'git')
source=("$pkgname-$pkgver.tar.gz::https://github.com/Vedant9500/WTF/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')  # Update this with actual checksum after creating the release

build() {
    cd "$srcdir/WTF-$pkgver"
    export CGO_ENABLED=0
    export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external"
    go build -ldflags "-X github.com/Vedant9500/WTF/internal/version.Version=$pkgver -X github.com/Vedant9500/WTF/internal/version.GitHash=release -X github.com/Vedant9500/WTF/internal/version.Build=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o wtf ./cmd/wtf
}

check() {
    cd "$srcdir/WTF-$pkgver"
    go test ./...
}

package() {
    cd "$srcdir/WTF-$pkgver"
    install -Dm755 wtf "$pkgdir/usr/bin/wtf"
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    install -Dm644 CHANGELOG.md "$pkgdir/usr/share/doc/$pkgname/CHANGELOG.md"
    install -Dm644 assets/commands.yml "$pkgdir/usr/share/wtf/commands.yml"
}
