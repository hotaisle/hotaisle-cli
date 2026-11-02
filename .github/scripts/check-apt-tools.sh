#!/usr/bin/env bash
set -euo pipefail

test_dir="$(mktemp -d)"
trap 'rm -rf "$test_dir"' EXIT
mkdir -p "$test_dir/repo/conf" "$test_dir/package/DEBIAN"

cat > "$test_dir/repo/conf/distributions" <<'EOF'
Codename: stable
Components: main
Architectures: amd64
Limit: 2
EOF

for version in 1 2 3; do
  cat > "$test_dir/package/DEBIAN/control" <<EOF
Package: hotaisle-apt-smoke
Version: 0.0.$version
Architecture: amd64
Section: utils
Priority: optional
Maintainer: Hot Aisle Tests <tests@example.invalid>
Description: Verify APT repository version retention
EOF
  dpkg-deb --root-owner-group --build "$test_dir/package" "$test_dir/$version.deb"
  reprepro --basedir "$test_dir/repo" includedeb stable "$test_dir/$version.deb"
done

versions="$(reprepro --basedir "$test_dir/repo" list stable hotaisle-apt-smoke | awk '{print $3}' | sort -V)"
if [[ "$versions" != $'0.0.2\n0.0.3' ]]; then
  printf 'Expected the two newest package versions, got:\n%s\n' "$versions" >&2
  exit 1
fi
echo 'APT tooling retains the configured number of package versions.'
