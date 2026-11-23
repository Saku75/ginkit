// Package version holds the build version. It is set at build time with
// -ldflags "-X ginkit/api/internal/version.Version=..."; scripts/build-images.sh
// gives both images the same value, and the web app's readiness probe checks it.
package version

var Version = "dev"
