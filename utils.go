package gocomplain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mjwhitta/log"
)

func execute(cmd []string) (string, error) {
	var b []byte
	var e error
	var tmp string

	if len(cmd) == 0 {
		return "", nil
	}

	if Debug {
		log.Debugf("%s", strings.Join(cmd, " "))
	}

	if b, e = exec.Command(cmd[0], cmd[1:]...).Output(); e != nil {
		switch e := e.(type) {
		case *exec.ExitError:
			tmp = strings.TrimSpace(string(e.Stderr))
			if tmp != "" {
				return "", fmt.Errorf(tmp)
			}
		default:
			return "", fmt.Errorf("failed to read cmd output: %w", e)
		}
	}

	return strings.TrimSuffix(string(b), "\n"), nil
}

func info(str string) {
	if !Quiet {
		log.Info(str)
	}
}

func runOutput(cmd []string, onlyWarn bool, skip ...int) {
	var cwd string
	var e error
	var isErr bool
	var out string
	var stdout string
	var trim []string

	if len(skip) == 0 {
		skip = []int{-1}
	}

	if stdout, e = execute(cmd); e != nil {
		isErr = true
		out = e.Error()
	} else {
		out = stdout
	}

	// Exit, if no usable output
	if out == "" {
		return
	} else if rIgnoredErr.MatchString(out) {
		return
	}

	cwd, _ = os.Getwd()

	for i, ln := range strings.Split(out, "\n") {
		if i == skip[0] {
			continue
		}

		// Clean up output
		trim = []string{"vet: ", "." + string(filepath.Separator)}
		if cwd != "" {
			trim = append(trim, cwd+string(filepath.Separator))
		}

		for _, prefix := range trim {
			ln = strings.TrimPrefix(ln, prefix)
		}

		if !onlyWarn && isErr {
			log.Err(ln)
		} else {
			log.Warn(ln)
		}
	}
}

func subInfof(str string, args ...any) {
	if !Quiet {
		log.SubInfof(str, args...)
	}
}
