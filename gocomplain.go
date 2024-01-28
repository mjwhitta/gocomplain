package gocomplain

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mjwhitta/log"
	"github.com/mjwhitta/where"
)

// FindSrcFiles will recursively traverse the current working
// directory and return a list of Go source files.
func FindSrcFiles(
	prune ...string,
) (map[string][]string, map[string][]string) {
	var dir string
	var src map[string][]string = map[string][]string{}
	var tests map[string][]string = map[string][]string{}

	filepath.WalkDir(
		".",
		func(fn string, d fs.DirEntry, e error) error {
			if e != nil {
				return nil
			}

			if d.IsDir() {
				for _, p := range prune {
					if p == d.Name() {
						return filepath.SkipDir
					}
				}

				return nil
			}

			dir = filepath.Dir(fn)
			fn = d.Name()

			for _, p := range prune {
				if p == fn {
					return nil
				}
			}

			if strings.HasSuffix(fn, "_test.go") {
				tests[dir] = append(tests[dir], fn)
			} else if strings.HasSuffix(fn, ".go") {
				src[dir] = append(src[dir], fn)
			}

			return nil
		},
	)

	return src, tests
}

// GoCyclo will analyze the provided Go source files for any functions
// that are overly complex.
func GoCyclo(over uint, src ...map[string][]string) {
	var cmd []string
	var e error
	var o string = strconv.Itoa(int(over))
	var stdout string

	log.Info("Checking code complexity (gocyclo)...")

	if len(src) == 0 {
		return
	}

	for i := range src {
		for dir, files := range src[i] {
			cmd = []string{"gocyclo", "--over", o}
			for _, fn := range files {
				cmd = append(cmd, filepath.Join(dir, fn))
			}

			if stdout, e = execute(cmd); e != nil {
				log.Err(e.Error())
			} else if stdout != "" {
				for _, ln := range strings.Split(stdout, "\n") {
					log.Warn(ln)
				}
			}
		}
	}
}

// GoFmt will format and simplify all Go source files.
func GoFmt() {
	var cmd []string = []string{"gofmt", "-l", "-s", "-w", "."}
	var e error
	var stdout string

	log.Info("Formatting code (gofmt)...")

	if stdout, e = execute(cmd); e != nil {
		log.Err(e.Error())
	} else if stdout != "" {
		for _, ln := range strings.Split(stdout, "\n") {
			log.Warn(ln)
		}
	}
}

// GoFumpt will format and optimize all Go source files.
func GoFumpt() {
	var cmd []string = []string{"gofumpt", "-e", "-l", "-w", "."}
	var e error
	var stdout string

	log.Info("Optimizing code (gofumpt)...")

	if stdout, e = execute(cmd); e != nil {
		log.Err(e.Error())
	} else if stdout != "" {
		for _, ln := range strings.Split(stdout, "\n") {
			log.Warn(ln)
		}
	}
}

// GoLint will lint all packages.
func GoLint(minConf float64) {
	var c string = strconv.FormatFloat(minConf, 'f', -1, 64)
	var cmd []string
	var e error
	var stdout string

	log.Info("Linting code (golint)...")

	if minConf == 0.8 {
		cmd = []string{"golint", "./..."}
	} else {
		cmd = []string{"golint", "-min_confidence", c, "./..."}
	}

	if stdout, e = execute(cmd); e != nil {
		if !nonModule.MatchString(e.Error()) {
			log.Err(e.Error())
		}
	} else if stdout != "" {
		for _, ln := range strings.Split(stdout, "\n") {
			log.Warn(ln)
		}
	}
}

// GoVet will vet all packages.
func GoVet(src ...map[string][]string) {
	var cmd []string

	log.Info("Vetting code (go vet)...")

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"go", "vet"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				runOutput(cmd, 0)
			}
		}
	} else {
		runOutput([]string{"go", "vet", "./..."}, 0)
	}
}

// IneffAssign will analyze all packages for any inefficient variable
// assignments.
func IneffAssign(src ...map[string][]string) {
	var cmd []string

	log.Info("Looking for inefficient assignments (ineffassign)...")

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"ineffassign"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				runOutput(cmd)
			}
		}
	} else {
		runOutput([]string{"ineffassign", "./..."})
	}
}

// LineLength will analyze the provided Go files for lines that are
// longer than the provided threshold.
func LineLength(threshold uint, src ...map[string][]string) {
	var e error
	var f *os.File
	var line string
	var lno int
	var s *bufio.Scanner

	log.Info("Checking for improper line-length...")

	for i := range src {
		for dir, files := range src[i] {
			for _, fn := range files {
				fn = filepath.Join(dir, fn)

				// Open file
				if f, e = os.Open(fn); e != nil {
					log.Errf("failed to open %s to read: %s", fn, e)
					continue
				}

				// Read file
				lno = 0
				s = bufio.NewScanner(f)
				for s.Scan() {
					line = strings.ReplaceAll(s.Text(), "\t", "    ")
					lno++

					if generated.MatchString(line) {
						break
					}

					if strings.HasPrefix(line, "//go:") {
						continue
					}

					if ll := len([]rune(line)); ll > int(threshold) {
						log.Warnf("%s:%d (%d) %s", fn, lno, ll, line)
					}
				}

				if e = s.Err(); e != nil {
					log.Errf("failed to read %s: %s", fn, e)
				}

				f.Close()
			}
		}
	}
}

// Misspell will look for spelling errors in proviced Go source files.
func Misspell(src ...map[string][]string) {
	var cmd []string
	var e error
	var stdout string

	log.Info("Checking spelling (misspell)...")

	for i := range src {
		for dir, files := range src[i] {
			cmd = []string{"misspell"}
			for _, fn := range files {
				cmd = append(cmd, filepath.Join(dir, fn))
			}

			if stdout, e = execute(cmd); e != nil {
				log.Err(e.Error())
			} else if stdout != "" {
				for _, ln := range strings.Split(stdout, "\n") {
					log.Warn(ln)
				}
			}
		}
	}
}

// SpellCheck will run the appropriate tool for the current OS and
// check for spelling errors in the provided Go source files.
func SpellCheck(
	ignore []string, skip []string, src ...map[string][]string,
) {
	var cmd []string

	log.Info("Checking spelling (codespell)...")

	switch runtime.GOOS {
	case "darwin", "linux":
		if where.Is("codespell") == "" {
			log.Err("codespell not found in PATH")
			return
		}

		cmd = []string{"codespell", "-d", "-f"}
		if len(ignore) > 0 {
			cmd = append(cmd, "-L", strings.Join(ignore, ","))
		}
		skip = append(skip, ".git", "*.pem", "go.mod", "go.sum")
		cmd = append(cmd, "-S", strings.Join(skip, ","))

		runOutput(cmd)
	// case "windows":
	// TODO find spellcheck tool for windows (codespell?)
	default:
		log.Errf("unsupported OS: %s", runtime.GOOS)
	}
}

// StaticCheck will perform static analysis on all packages.
func StaticCheck(src ...map[string][]string) {
	var cmd []string

	log.Info("Running static analysis (staticcheck)...")

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"staticcheck"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				runOutput(cmd, 0)
			}
		}
	} else {
		runOutput([]string{"staticcheck", "./..."}, 0)
	}
}

// UpdateInstall will install the newest versions of the underlying
// tools.
func UpdateInstall() {
	var cmd []string = []string{
		"go",
		"install",
		"--buildvcs=false",
		"--ldflags=-s -w",
		"--trimpath",
	}
	var e error
	var found bool
	var stdout string
	var tools map[string]string = map[string]string{
		"gocyclo":     "github.com/fzipp/gocyclo/cmd/gocyclo",
		"gofumpt":     "mvdan.cc/gofumpt",
		"golint":      "golang.org/x/lint/golint",
		"ineffassign": "github.com/gordonklaus/ineffassign",
		"misspell":    "github.com/client9/misspell/cmd/misspell",
		"staticcheck": "honnef.co/go/tools/cmd/staticcheck",
	}

	log.Info("Installing newest versions of each tool...")
	for name, tool := range tools {
		log.SubInfof("%s...", name)
		stdout, e = execute(append(cmd, tool+"@latest"))
		if e != nil {
			log.Err(e.Error())
		} else if stdout != "" {
			for _, ln := range strings.Split(stdout, "\n") {
				log.Warn(ln)
			}
		}
	}

	switch runtime.GOOS {
	case "darwin":
		if where.Is("codespell") == "" {
			log.Warn("Please run \"brew install codespell\".")
		}
	case "linux":
		if where.Is("codespell") != "" {
			return
		}

		for _, pkgMgr := range pkgMgrs {
			if where.Is(pkgMgr[0]) != "" {
				found = true
				log.Warnf("Please run \"%s\".", pkgMgr[1])
				break
			}
		}

		if !found {
			log.Warn("Unknown package manager.")
			log.Warn("Please install codespell.")
		}
	case "windows":
		// TODO install spellcheck tool for windows
		log.Warnf(
			"spellcheck tool for %s not implemented",
			runtime.GOOS,
		)
	}
}
