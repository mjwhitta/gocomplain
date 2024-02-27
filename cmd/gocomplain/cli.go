package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/mjwhitta/cli"
	"github.com/mjwhitta/gocomplain"
	hl "github.com/mjwhitta/hilighter"
	"github.com/mjwhitta/log"
)

// Exit status
const (
	Good = iota
	InvalidOption
	MissingOption
	InvalidArgument
	MissingArgument
	ExtraArgument
	Exception
)

// Flags
var flags struct {
	confidence float64
	debug      bool
	ignore     cli.StringList
	length     uint
	nocolor    bool
	over       uint
	prune      cli.StringList
	quiet      bool
	skip       cli.StringList
	verbose    bool
	version    bool
}

func init() {
	// Configure cli package
	cli.Align = true
	cli.Authors = []string{"Miles Whittaker <mj@whitta.dev>"}
	cli.Banner = hl.Sprintf(
		"%s [OPTIONS] [action1]... [actionN]",
		os.Args[0],
	)
	cli.BugEmail = "gocomplain.bugs@whitta.dev"
	cli.ExitStatus(
		"Normally the exit status is 0. In the event of an error the",
		"exit status will be one of the below:\n\n",
		hl.Sprintf("  %d: Invalid option\n", InvalidOption),
		hl.Sprintf("  %d: Missing option\n", MissingOption),
		hl.Sprintf("  %d: Invalid argument\n", InvalidArgument),
		hl.Sprintf("  %d: Missing argument\n", MissingArgument),
		hl.Sprintf("  %d: Extra argument\n", ExtraArgument),
		hl.Sprintf("  %d: Exception", Exception),
	)
	cli.Info(
		"GoComplain combines multiple other Go source analyzing",
		"tools. Currently supported functionality includes: gocyclo,",
		"gofmt, gofumpt, golint, go vet, ineffassign, line-length",
		"verification, spellcheck, and staticcheck. The spellcheck",
		"functionality uses the misspell Go module as well as",
		"codespell on Linux and macOS.",
	)
	cli.Section(
		"ACTIONS - COMMANDS",
		"h, help\nDisplay this help message.\n\n",
		"i, install, u, update, upgrade\n",
		"Install or reinstall underlying tools.\n\n",
		"v, version\nShow version.",
	)
	cli.Section(
		"ACTIONS - ENV",
		"allos\nCheck all supported GOOS.\n\n",
		"darwin, linux, windows (default: "+runtime.GOOS+")\n",
		"Check the specified GOOS.\n\n",
		"ao, d, l, w\nShorthand for associated GOOS.",
	)
	cli.Section(
		"ACTIONS - TOOLS",
		"alltools (default)\nRun all tools.\n\n",
		"gocyclo, gofmt, gofumpt, golint, govet, ineffassign,",
		"line-length, spellcheck, staticheck\n",
		"Run the specified tool.\n\n",
		"at, cyclo, fmt, fumpt, lint, vet, ineff, ll, spell, static",
		"\n",
		"Shorthand for associated tools.",
	)
	cli.SeeAlso = []string{
		"codespell",
		"go vet",
		"gocyclo",
		"gofmt",
		"gofumpt",
		"golint",
		"ineffassign",
		"misspell",
		"staticcheck",
	}
	cli.Title = "GoComplain"

	// Parse cli flags
	cli.Flag(
		&flags.confidence,
		"c",
		"confidence",
		0.8,
		"Only complain about golint problems with specified minimum",
		"confidence (default: 0.8).",
	)
	cli.Flag(
		&flags.debug,
		"d",
		"debug",
		false,
		"Enable printing of executed sub-processes.",
		true,
	)
	cli.Flag(
		&flags.ignore,
		"i",
		"ignore",
		"Ignore words when checking spelling.",
	)
	cli.Flag(
		&flags.length,
		"l",
		"length",
		70,
		"Set max length of source code lines (default: 70).",
	)
	cli.Flag(
		&flags.nocolor,
		"no-color",
		false,
		"Disable colorized output.",
	)
	cli.Flag(
		&flags.over,
		"o",
		"over",
		15,
		"Only complain about functions over specified complexity",
		"(default: 15).",
	)
	cli.Flag(
		&flags.prune,
		"p",
		"prune",
		"Prune directories/files when analyzing source files.",
	)
	cli.Flag(
		&flags.quiet,
		"q",
		"quiet",
		false,
		"Hide information log messages.",
	)
	cli.Flag(
		&flags.skip,
		"s",
		"skip",
		"Skip directories/files when checking spelling.",
	)
	cli.Flag(
		&flags.verbose,
		"v",
		"verbose",
		false,
		"Show stacktrace, if error.",
	)
	cli.Flag(&flags.version, "V", "version", false, "Show version.")
	cli.Parse()
}

// Process cli flags and ensure no issues
func validate() {
	var tmp []string

	hl.Disable(flags.nocolor)

	for _, arg := range cli.Args() {
		switch arg {
		case "h", "help":
			cli.Usage(0)
		case "i", "install", "u", "update", "upgrade":
			if cli.NArg() != 1 {
				cli.Usage(ExtraArgument)
			}
		case "v", "version":
			flags.version = true
		}
	}

	// Validate cli flags
	if flags.length < 70 {
		log.ErrX(InvalidOption, "Less than 70? Who hurt you?")
	} else if flags.length > 100 {
		log.ErrX(InvalidOption, "Greater than 100? You monster!")
	}

	// Short circuit, if version was requested
	if flags.version {
		hl.Printf("gocomplain version %s\n", gocomplain.Version)
		os.Exit(Good)
	}

	// Fix string lists
	for _, ignore := range flags.ignore {
		tmp = append(tmp, strings.Split(ignore, ",")...)
	}
	flags.ignore = tmp

	tmp = []string{}
	for _, prune := range flags.prune {
		tmp = append(tmp, strings.Split(prune, ",")...)
	}
	flags.prune = tmp

	tmp = []string{}
	for _, skip := range flags.skip {
		tmp = append(tmp, strings.Split(skip, ",")...)
	}
	flags.skip = tmp
}
