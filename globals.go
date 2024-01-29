package gocomplain

import (
	"regexp"
	"strings"
)

// Debug will turn on printing of executed sub-processes.
var Debug bool

var (
	generated *regexp.Regexp = regexp.MustCompile(
		`^//\sCode\sgenerated\s.*\sDO\sNOT\sEDIT\.$`,
	)
	ignoredErrs []string = []string{
		"does not contain main module",
		"matched no packages",
		"-buildvcs=false to disable VCS stamping.",
	}
	rIgnoredErr *regexp.Regexp = regexp.MustCompile(
		strings.Join(ignoredErrs, "|"),
	)
)

var pkgMgrs = [][]string{
	// Alpine
	{"apk", "sudo apk add py3-codespell"},
	// Arch
	{"yay", "yay -S codespell"},
	{"pacman", "sudo pacman -S codespell"},
	// Debian
	{"apt-get", "sudo apt-get install codespell"},
	{"apt", "sudo apt install codespell"},
	// OpenSUSE
	{"zypper", "sudo zypper in codespell"},
	// RedHat
	{"dnf", "sudo dnf install codespell"},
	{"yum", "sudo yum install codespell"},
}

// Version is the package version.
const Version = "0.4.2"
