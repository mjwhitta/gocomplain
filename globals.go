package gocomplain

import "regexp"

// Debug will turn on printing of executed sub-processes.
var Debug bool

var (
	generated *regexp.Regexp = regexp.MustCompile(
		`^//\sCode\sgenerated\s.*\sDO\sNOT\sEDIT\.$`,
	)
	nonModule *regexp.Regexp = regexp.MustCompile(
		`does not contain main module|matched no packages`,
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
const Version = "0.4.1"
