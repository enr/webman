//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

package utils

func executablePath(pkg string) string {
	return pkg
}
