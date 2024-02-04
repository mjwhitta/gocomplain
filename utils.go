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

func run(cmd []string) []string {
	var cwd string
	var e error
	var out []string
	var stdout string
	var trim []string

	if stdout, e = execute(cmd); e != nil {
		stdout = e.Error()
	}

	// Exit, if no usable output
	if stdout == "" {
		return nil
	} else if rIgnoredErr.MatchString(stdout) {
		return nil
	}

	cwd, _ = os.Getwd()

	for _, ln := range strings.Split(stdout, "\n") {
		// Clean up output
		trim = []string{"vet: ", "." + string(filepath.Separator)}
		if cwd != "" {
			trim = append(trim, cwd+string(filepath.Separator))
		}

		for _, prefix := range trim {
			ln = strings.TrimPrefix(ln, prefix)
		}

		out = append(out, ln)
	}

	return out
}

func subInfof(str string, args ...any) {
	if !Quiet {
		log.SubInfof(str, args...)
	}
}
