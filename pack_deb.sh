#!/usr/bin/env bash
set -ex

# Define variables
STAGING_DIR="tmp_deb"
DEBIAN_DIR="${STAGING_DIR}/DEBIAN"

# Clean up any existing staging dir
rm -rf "${STAGING_DIR}"
mkdir -p "${STAGING_DIR}"

# Create standard directories
mkdir -p "${DEBIAN_DIR}"
mkdir -p "${STAGING_DIR}/usr/bin"
mkdir -p "${STAGING_DIR}/usr/lib/systemd/system"
mkdir -p "${STAGING_DIR}/usr/lib/systemd/user"
mkdir -p "${STAGING_DIR}/usr/share/icons/hicolor/512x512/apps"
mkdir -p "${STAGING_DIR}/usr/share/applications"
mkdir -p "${STAGING_DIR}/etc/default"

# Copy binaries
cp v2raya "${STAGING_DIR}/usr/bin/v2raya"
cp v2raya_core "${STAGING_DIR}/usr/bin/v2raya_core"

# Copy config and service files
cp install/universal/v2raya.service "${STAGING_DIR}/usr/lib/systemd/system/v2raya.service"
cp install/universal/v2raya-lite.service "${STAGING_DIR}/usr/lib/systemd/user/v2raya-lite.service"
cp install/universal/v2raya.png "${STAGING_DIR}/usr/share/icons/hicolor/512x512/apps/v2raya.png"
cp install/universal/v2raya.desktop "${STAGING_DIR}/usr/share/applications/v2raya.desktop"
cp install/universal/v2raya.default "${STAGING_DIR}/etc/default/v2raya"

# Set permissions
chmod 755 "${STAGING_DIR}/usr/bin/v2raya"
chmod 755 "${STAGING_DIR}/usr/bin/v2raya_core"

# Write DEBIAN/control file
cat <<EOF > "${DEBIAN_DIR}/control"
Package: v2raya
Version: 2.2.5-test1
Section: net
Priority: optional
Architecture: arm64
Maintainer: v2raya <v2raya@v2raya.org>
Depends: ca-certificates
Description: A web GUI client of Project V which supports V2Ray, Xray, Trojan, Shadowsocks, etc.
EOF

# Write DEBIAN/postinst script
cat <<EOF > "${DEBIAN_DIR}/postinst"
#!/usr/bin/env bash
systemctl daemon-reload
echo -e "\033[36m******************************\033[0m"
echo -e "\033[36m*      v2rayA Installed!     *\033[0m"
echo -e "\033[36m******************************\033[0m"
EOF
chmod 755 "${DEBIAN_DIR}/postinst"

# Write DEBIAN/prerm script
cat <<EOF > "${DEBIAN_DIR}/prerm"
#!/usr/bin/env bash
systemctl stop v2raya || true
systemctl disable v2raya || true
systemctl daemon-reload
EOF
chmod 755 "${DEBIAN_DIR}/prerm"

# Build the Debian package
dpkg-deb --root-owner-group -b "${STAGING_DIR}" v2raya_2.2.5-test1_arm64.deb

# Cleanup staging directory
rm -rf "${STAGING_DIR}"

echo "Debian package successfully built: v2raya_2.2.5-test1_arm64.deb"
