package gocomplain

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	hl "github.com/mjwhitta/hilighter"
	"github.com/mjwhitta/log"
	"github.com/mjwhitta/where"
)

// FindSrcFiles will recursively traverse the provided directory and
// return a list of Go source files.
func FindSrcFiles(
	search string, prune ...string,
) (map[string][]string, map[string][]string, map[string][]string) {
	var other map[string][]string = map[string][]string{}
	var src map[string][]string = map[string][]string{}
	var tests map[string][]string = map[string][]string{}

	_ = filepath.WalkDir(
		search,
		func(fn string, d fs.DirEntry, e error) error {
			var dir string

			if e != nil {
				return nil
			}

			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}

				if slices.Contains(prune, d.Name()) {
					return filepath.SkipDir
				}

				return nil
			}

			dir = filepath.Dir(fn)
			fn = d.Name()

			if alwaysIgnore.MatchString(fn) {
				return nil
			}

			if slices.Contains(prune, fn) {
				return nil
			}

			if strings.HasSuffix(fn, "_test.go") {
				tests[dir] = append(tests[dir], fn)
			} else if strings.HasSuffix(fn, ".go") {
				src[dir] = append(src[dir], fn)
			} else {
				other[dir] = append(other[dir], fn)
			}

			return nil
		},
	)

	return src, tests, other
}

// GoCyclo will analyze the provided Go source files for any functions
// that are overly complex.
func GoCyclo(over uint) []string {
	return run(
		[]string{"gocyclo", "--over", strconv.Itoa(int(over)), "."},
	)
}

// GoFmt will format and simplify all Go source files.
func GoFmt() []string {
	return run([]string{"gofmt", "-l", "-s", "-w", "."})
}

// GoFumpt will format and optimize all Go source files.
func GoFumpt() []string {
	return run([]string{"gofumpt", "-e", "-l", "-w", "."})
}

// GoLint will lint all packages.
func GoLint(minConf float64) []string {
	var c string = strconv.FormatFloat(minConf, 'f', -1, 64)
	var cmd []string = []string{"golint"}

	if minConf != 0.8 {
		cmd = append(cmd, "-min_confidence", c)
	}

	cmd = append(cmd, "./...")

	return run(cmd)
}

// GoVet will vet all packages.
func GoVet(src ...map[string][]string) []string {
	var cmd []string
	var out []string

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"go", "vet"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				out = append(out, run(cmd)...)
			}
		}

		return out
	}

	return run([]string{"go", "vet", "./..."})
}

// IneffAssign will analyze all packages for any inefficient variable
// assignments.
func IneffAssign(src ...map[string][]string) []string {
	var cmd []string
	var out []string

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"ineffassign"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				out = append(out, run(cmd)...)
			}
		}

		return out
	}

	return run([]string{"ineffassign", "./..."})
}

// LineLength will analyze the provided Go files for lines that are
// longer than the provided threshold.
func LineLength(threshold uint, src ...map[string][]string) []string {
	var e error
	var f *os.File
	var line string
	var lno int
	var out []string
	var s *bufio.Scanner

	for i := range src {
		for dir, files := range src[i] {
			for _, fn := range files {
				fn = filepath.Join(dir, fn)

				// Open file
				if f, e = os.Open(fn); e != nil {
					out = append(
						out,
						hl.Sprintf("failed to read %s: %s", fn, e),
					)
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

					if structTags.MatchString(line) {
						continue
					}

					if ll := len([]rune(line)); ll > int(threshold) {
						out = append(
							out,
							hl.Sprintf(
								"%s:%d (%d) %s",
								fn,
								lno,
								ll,
								line,
							),
						)
					}
				}

				if e = s.Err(); e != nil {
					out = append(
						out,
						hl.Sprintf("failed to read %s: %s", fn, e),
					)
				}

				f.Close()
			}
		}
	}

	return out
}

// Misspell will look for spelling errors in provided Go source files.
func Misspell(ignore []string, src ...map[string][]string) []string {
	var cmd []string = []string{"misspell"}
	var out []string
	var tmp []string

	if len(ignore) > 0 {
		cmd = append(cmd, "-i", strings.Join(ignore, ","))
	}

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				tmp = []string{}

				for _, file := range files {
					tmp = append(tmp, filepath.Join(dir, file))
				}

				out = append(out, run(append(cmd, tmp...))...)
			}
		}

		return out
	}

	return run(append(cmd, "."))
}

// SpellCheck will run the appropriate tool for the current OS and
// check for spelling errors in the provided Go source files.
func SpellCheck(
	ignore []string, skip []string, src ...map[string][]string,
) []string {
	var cmd []string

	switch runtime.GOOS {
	case "darwin", "linux":
		if where.Is("codespell") == "" {
			return []string{"codespell not found in PATH"}
		}

		cmd = []string{"codespell", "-d", "-f"}
		if len(ignore) > 0 {
			cmd = append(
				cmd,
				"-L",
				strings.ToLower(strings.Join(ignore, ",")),
			)
		}

		skip = append(
			skip,
			".git*",
			"*.db",
			"*.der",
			"*.dll",
			"*.exe",
			"*.drawio",
			"*.exe",
			"*.gif",
			"*.gz",
			"*.jar",
			"*.jpeg",
			"*.jpg",
			"*.pdf",
			"*.pem",
			"*.png",
			"*.so",
			"*.tar",
			"*.tgz",
			"*.xz",
			"*.zip",
			"go.mod",
			"go.sum",
		)
		cmd = append(cmd, "-S", strings.Join(skip, ","))

		return run(cmd)
	// case "windows":
	// TODO find spellcheck tool for windows (codespell?)
	default:
		return []string{
			hl.Sprintf("unsupported OS: %s", runtime.GOOS),
		}
	}
}

// StaticCheck will perform static analysis on all packages.
func StaticCheck(src ...map[string][]string) []string {
	var cmd []string
	var out []string

	if len(src) > 0 {
		for i := range src {
			for dir, files := range src[i] {
				cmd = []string{"staticcheck"}
				for _, file := range files {
					cmd = append(cmd, filepath.Join(dir, file))
				}

				out = append(out, run(cmd)...)
			}
		}

		return out
	}

	return run([]string{"staticcheck", "./..."})
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

	info("Installing newest versions of each tool...")
	for _, tool := range tools {
		subInfof("%s...", tool[0])
		run(append(cmd, tool[1]+"@latest"))
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
