package gocomplain

import (
	"regexp"
	"strings"
)

// Debug will turn on debug log messages.
var Debug bool

var (
	alwaysIgnore *regexp.Regexp = regexp.MustCompile("" +
		`\.git*|.*\.(` +
		`db|der|dll|drawio|exe|gif|gz|jar|jpeg|jpg|pdf|pem|png|so` +
		`tar|tgz|xz|zip` +
		`)`,
	)
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

// Quiet can be used to disable information log messages.
var Quiet bool

// Version is the package version.
const Version string = "0.8.0"
