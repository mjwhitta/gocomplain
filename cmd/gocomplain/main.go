package main

import (
	"os"
	"path/filepath"

	"github.com/mjwhitta/cli"
	"github.com/mjwhitta/gocomplain"
	"github.com/mjwhitta/log"
	"github.com/mjwhitta/pathname"
)

var all []string = []string{
	"gofmt",
	"gofumpt",
	"gocyclo",
	"ineffassign",
	"golint",
	"govet",
	"staticcheck",
	"spellcheck",
	"line-length",
}

var inMod bool

func isCmd(arg string) bool {
	switch arg {
	case "i", "install", "u", "update", "upgrade":
		gocomplain.UpdateInstall()
		return true
	}

	return false
}

func isEnv(arg string) (bool, []string) {
	switch arg {
	case "d", "darwin":
		return true, []string{"darwin"}
	case "l", "linux":
		return true, []string{"linux"}
	case "w", "windows":
		return true, []string{"windows"}
	}

	return false, nil
}

func isTool(arg string) (bool, []string) {
	switch arg {
	case "a", "all":
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
	var tools []string

	validate()

	gocomplain.Debug = flags.debug

	if inMod, e = setup(); e != nil {
		panic(e)
	}

	if cli.NArg() == 0 {
		tools = all
	}

	for _, arg := range cli.Args() {
		if ok := isCmd(arg); ok {
			// Do nothing
		} else if ok, add := isEnv(arg); ok {
			tools = append(tools, add...)
		} else if ok, add := isTool(arg); ok {
			tools = append(tools, add...)
		} else {
			cli.Usage(InvalidArgument)
		}
	}

	if len(tools) > 0 {
		src, tests = gocomplain.FindSrcFiles(flags.prune...)
		run(tools, src, tests)
	}

	log.Good("Done")
}

func run(tools []string, src ...map[string][]string) {
	for _, tool := range tools {
		switch tool {
		case "darwin", "linux", "windows":
			os.Setenv("GOOS", tool)
		case "gocyclo":
			gocomplain.GoCyclo(flags.over)
		case "gofmt":
			gocomplain.GoFmt()
		case "gofumpt":
			gocomplain.GoFumpt()
		case "golint":
			gocomplain.GoLint(flags.confidence)
		case "govet":
			if inMod {
				gocomplain.GoVet()
			} else {
				gocomplain.GoVet(src...)
			}
		case "ineffassign":
			if inMod {
				gocomplain.IneffAssign()
			} else {
				gocomplain.IneffAssign(src...)
			}
		case "line-length":
			gocomplain.LineLength(flags.length, src...)
		case "spellcheck":
			gocomplain.Misspell()
			gocomplain.SpellCheck(flags.ignore, flags.skip, src...)
		case "staticcheck":
			if inMod {
				gocomplain.StaticCheck()
			} else {
				gocomplain.StaticCheck(src...)
			}
		}
	}
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
