package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mjwhitta/cli"
	"github.com/mjwhitta/gocomplain"
	"github.com/mjwhitta/log"
	"github.com/mjwhitta/pathname"
)

var (
	all []string = []string{
		"gofmt",
		"gofumpt",
		"gocyclo",
		"ineffassign",
		"golint",
		"govet",
		"line-length",
		"spellcheck",
		"staticcheck",
	}
	inMod bool
	oses  []string
	rm    []string
	tools []string
)

func infof(str string, args ...any) {
	if !flags.quiet {
		log.Infof(str, args...)
	}
}

func isCmd(arg string) bool {
	switch arg {
	case "i", "install", "u", "update", "upgrade":
		gocomplain.UpdateInstall()
		return true
	}

	return false
}

func isOS(arg string) (bool, []string) {
	switch arg {
	case "d", "darwin":
		return true, []string{"darwin"}
	case "dlw":
		return true, []string{"darwin", "linux", "windows"}
	case "l", "linux":
		return true, []string{"linux"}
	case "w", "windows":
		return true, []string{"windows"}
	}

	return false, nil
}

func isRemove(arg string) (bool, string) {
	if !strings.HasPrefix(arg, "no") {
		return false, ""
	}

	arg = strings.TrimPrefix(arg, "no")

	switch arg {
	case "cyclo", "gocyclo":
		return true, "gocyclo"
	case "fmt", "gofmt":
		return true, "gofmt"
	case "fumpt", "gofumpt":
		return true, "gofumpt"
	case "golint", "lint":
		return true, "golint"
	case "govet", "vet":
		return true, "govet"
	case "ineff", "ineffassign":
		return true, "ineffassign"
	case "ll", "line-length":
		return true, "line-length"
	case "spell", "spellcheck":
		return true, "spellcheck"
	case "static", "staticcheck":
		return true, "staticcheck"
	}

	return false, ""
}

func isTool(arg string) (bool, []string) {
	switch arg {
	case "all":
		return true, all
	case "cyclo", "gocyclo":
		return true, []string{"gocyclo"}
	case "fmt", "gofmt":
		return true, []string{"gofmt"}
	case "fumpt", "gofumpt":
		return true, []string{"gofumpt"}
	case "golint", "lint":
		return true, []string{"golint"}
	case "govet", "vet":
		return true, []string{"govet"}
	case "ineff", "ineffassign":
		return true, []string{"ineffassign"}
	case "ll", "line-length":
		return true, []string{"line-length"}
	case "spell", "spellcheck":
		return true, []string{"spellcheck"}
	case "static", "staticcheck":
		return true, []string{"staticcheck"}
	}

	return false, nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			if flags.verbose {
				panic(r.(error).Error())
			}
			log.ErrX(Exception, r.(error).Error())
		}
	}()

	var e error
	var src map[string][]string
	var tests map[string][]string

	validate()

	gocomplain.Debug = flags.debug
	gocomplain.Quiet = flags.quiet

	if inMod, e = setup(); e != nil {
		panic(e)
	}

	for _, arg := range cli.Args() {
		if ok := isCmd(arg); ok {
			os.Exit(Good)
		} else if ok, add := isOS(arg); ok {
			oses = append(oses, add...)
		} else if ok, add := isRemove(arg); ok {
			rm = append(rm, add)
		} else if ok, add := isTool(arg); ok {
			tools = append(tools, add...)
		} else {
			cli.Usage(InvalidArgument)
		}
	}

	if len(oses) == 0 {
		oses = append(oses, runtime.GOOS)
	}

	if len(tools) == 0 {
		tools = append(tools, all...)
	}

	for i := range rm {
		for j := range tools {
			if tools[j] == rm[i] {
				tools = append(tools[:j], tools[j+1:]...)
				break
			}
		}
	}

	src, tests = gocomplain.FindSrcFiles(".", flags.prune...)
	run(src, tests)

	if !flags.quiet {
		log.Good("Done")
	}
}

func output(out []string) {
	for _, ln := range out {
		log.Warn(ln)
	}
}

func run(src ...map[string][]string) {
	var lineLength bool
	var spellcheck bool

	for _, goos := range oses {
		infof("Setting GOOS to %s", goos)
		os.Setenv("GOOS", goos)

		if ll, spell := runOS(src...); ll && spell {
			lineLength = true
			spellcheck = true
		} else if ll {
			lineLength = true
		} else if spell {
			spellcheck = true
		}
	}

	if lineLength {
		infof("Checking for improper line-length...")
		output(gocomplain.LineLength(flags.length, src...))
	}

	if spellcheck {
		os.Setenv("GOOS", runtime.GOOS)

		infof("Checking spelling (misspell)...")
		output(gocomplain.Misspell(flags.ignore))

		infof("Checking spelling (codespell)...")
		output(
			gocomplain.SpellCheck(flags.ignore, flags.skip, src...),
		)
	}
}

func runOS(src ...map[string][]string) (bool, bool) {
	var lineLength bool
	var spellcheck bool

	for _, tool := range tools {
		switch tool {
		case "gocyclo":
			subInfof("Checking code complexity (gocyclo)...")
			output(gocomplain.GoCyclo(flags.over))
		case "gofmt":
			subInfof("Formatting code (gofmt)...")
			output(gocomplain.GoFmt())
		case "gofumpt":
			subInfof("Optimizing code (gofumpt)...")
			output(gocomplain.GoFumpt())
		case "golint":
			subInfof("Linting code (golint)...")
			output(gocomplain.GoLint(flags.confidence))
		case "govet":
			subInfof("Vetting code (go vet)...")
			if inMod {
				output(gocomplain.GoVet())
			} else {
				output(gocomplain.GoVet(src...))
			}
		case "ineffassign":
			subInfof(
				"Looking for inefficient assignments (%s)...",
				tool,
			)
			if inMod {
				output(gocomplain.IneffAssign())
			} else {
				output(gocomplain.IneffAssign(src...))
			}
		case "line-length":
			lineLength = true
		case "spellcheck":
			spellcheck = true
		case "staticcheck":
			subInfof("Running static analysis (staticcheck)...")
			if inMod {
				output(gocomplain.StaticCheck())
			} else {
				output(gocomplain.StaticCheck(src...))
			}
		}
	}

	return lineLength, spellcheck
}

func setup() (bool, error) {
	var cwd string
	var e error
	var mod string
	var tmp string

	if cwd, e = os.Getwd(); e != nil {
		return false, e
	}

	for {
		if flags.debug {
			log.Debugf("Checking %s for go.mod", cwd)
		}

		mod = filepath.Join(cwd, "go.mod")

		if ok, _ := pathname.DoesExist(mod); ok {
			break
		} else if tmp = filepath.Dir(cwd); tmp == cwd {
			return false, nil // Not in a module
		}

		cwd = tmp
	}

	if flags.debug {
		log.Debugf("Found %s", mod)
	}

	if e = os.Chdir(cwd); e != nil {
		return false, e
	}

	return true, nil
}

func subInfof(str string, args ...any) {
	if !flags.quiet {
		log.SubInfof(str, args...)
	}
}
