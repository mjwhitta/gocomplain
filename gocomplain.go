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
func GoCyclo(over uint) {
	var o string = strconv.Itoa(int(over))

	log.Info("Checking code complexity (gocyclo)...")
	runOutput([]string{"gocyclo", "--over", o, "."}, false)
}

// GoFmt will format and simplify all Go source files.
func GoFmt() {
	log.Info("Formatting code (gofmt)...")
	runOutput([]string{"gofmt", "-l", "-s", "-w", "."}, false)
}

// GoFumpt will format and optimize all Go source files.
func GoFumpt() {
	log.Info("Optimizing code (gofumpt)...")
	runOutput([]string{"gofumpt", "-e", "-l", "-w", "."}, false)
}

// GoLint will lint all packages.
func GoLint(minConf float64) {
	var c string = strconv.FormatFloat(minConf, 'f', -1, 64)
	var cmd []string = []string{"golint"}

	if minConf != 0.8 {
		cmd = append(cmd, "-min_confidence", c)
	}

	cmd = append(cmd, "./...")

	log.Info("Linting code (golint)...")
	runOutput(cmd, false)
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

				runOutput(cmd, true, 0)
			}
		}
	} else {
		runOutput([]string{"go", "vet", "./..."}, true, 0)
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

				runOutput(cmd, true)
			}
		}
	} else {
		runOutput([]string{"ineffassign", "./..."}, true)
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

// Misspell will look for spelling errors in provided Go source files.
func Misspell(ignore []string) {
	var cmd []string = []string{"misspell"}

	if len(ignore) > 0 {
		cmd = append(cmd, "-i", strings.Join(ignore, ","))
	}
	cmd = append(cmd, ".")

	log.Info("Checking spelling (misspell)...")
	runOutput(cmd, false)
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

		runOutput(cmd, true)
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

				runOutput(cmd, true, 0)
			}
		}
	} else {
		runOutput([]string{"staticcheck", "./..."}, true, 0)
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
	var found bool
	var tools [][]string = [][]string{
		{"gocyclo", "github.com/fzipp/gocyclo/cmd/gocyclo"},
		{"gofumpt", "mvdan.cc/gofumpt"},
		{"golint", "golang.org/x/lint/golint"},
		{"ineffassign", "github.com/gordonklaus/ineffassign"},
		{"misspell", "github.com/client9/misspell/cmd/misspell"},
		{"staticcheck", "honnef.co/go/tools/cmd/staticcheck"},
	}

	log.Info("Installing newest versions of each tool...")
	for _, tool := range tools {
		log.SubInfof("%s...", tool[0])
		runOutput(append(cmd, tool[1]+"@latest"), false)
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
