#!/usr/bin/env sh
# Builds both images with one version, so the web app's /readyz matches the API.
# Usage: scripts/build-images.sh [version]   (defaults to git describe)
set -eu
cd "$(dirname "$0")/.."

VERSION="${1:-$(git describe --tags --always --dirty)}"
docker build --build-arg "VERSION=$VERSION" -t "ginkit-api:$VERSION" apps/api
docker build --build-arg "VERSION=$VERSION" -f apps/web/Dockerfile -t "ginkit-web:$VERSION" .
echo "built ginkit-api:$VERSION and ginkit-web:$VERSION"
