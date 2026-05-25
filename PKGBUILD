pkgname=yummi
pkgver=0.1.0
pkgrel=1
pkgdesc='Self-hosted recipe manager with web UI, URL import, and OpenAI integration'
arch=('x86_64' 'aarch64')
url='https://github.com/carnager/yummi'
license=('MIT')
makedepends=('go')
backup=('etc/yummi/yummi.conf')
install=yummi.install
source=()

build() {
    cd "$srcdir/.."
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o yummi .
}

package() {
    cd "$srcdir/.."

    # Binary
    install -Dm755 yummi "$pkgdir/usr/bin/yummi"

    # Systemd service
    install -Dm644 yummi.service "$pkgdir/usr/lib/systemd/system/yummi.service"

    # Config file
    install -dm755 "$pkgdir/etc/yummi"
    cat > "$pkgdir/etc/yummi/yummi.conf" <<'EOF'
# Yummi configuration
# Source this file in the systemd override or set these environment variables.
YUMMI_PORT=:8080
YUMMI_DATA_DIR=/var/lib/yummi
#YUMMI_SECRET=
#YUMMI_OPENAI_KEY=
EOF
    chmod 640 "$pkgdir/etc/yummi/yummi.conf"

    # Data directory
    install -dm750 "$pkgdir/var/lib/yummi"

    # Tmpfiles for ownership
    install -Dm644 /dev/stdin "$pkgdir/usr/lib/tmpfiles.d/yummi.conf" <<'EOF'
d /var/lib/yummi 0750 yummi yummi -
EOF

    # Sysusers
    install -Dm644 /dev/stdin "$pkgdir/usr/lib/sysusers.d/yummi.conf" <<'EOF'
u yummi - "Yummi Recipe Manager" /var/lib/yummi
EOF
}
