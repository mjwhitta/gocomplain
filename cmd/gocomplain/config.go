package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	Confidence float64  `json:"confidence"`
	file       string   `json:"-"`
	Ignore     []string `json:"ignore"`
	Length     uint     `json:"length"`
	Over       uint     `json:"over"`
	Prune      []string `json:"prune"`
	Quiet      bool     `json:"quiet"`
	Skip       []string `json:"skip"`
}

var cfg *config

func init() {
	var b []byte
	var e error
	var fn string

	if fn, e = os.UserConfigDir(); e != nil {
		panic(fmt.Errorf("user has no cfg directory: %w", e))
	}

	fn = filepath.Join(fn, "gocomplain", "rc")
	b, e = os.ReadFile(fn)

	if (e != nil) || (len(bytes.TrimSpace(b)) == 0) {
		// Default cfg
		cfg = &config{
			Confidence: 0.8,
			file:       fn,
			Ignore:     []string{},
			Length:     70,
			Over:       15,
			Prune:      []string{},
			Skip:       []string{},
		}

		if e = cfg.Save(); e != nil {
			panic(e)
		}
	} else {
		if e = json.Unmarshal(b, &cfg); e != nil {
			panic(fmt.Errorf("invalid cfg: %w", e))
		}
	}

	if cfg.Confidence == 0 {
		cfg.Confidence = 0.8
	}

	if cfg.Ignore == nil {
		cfg.Ignore = []string{}
	}

	if cfg.Length == 0 {
		cfg.Length = 70
	}

	if cfg.Over == 0 {
		cfg.Over = 15
	}

	if cfg.Prune == nil {
		cfg.Prune = []string{}
	}

	if cfg.Skip == nil {
		cfg.Skip = []string{}
	}
}

func (c *config) Save() error {
	var e error

	if e = os.MkdirAll(filepath.Dir(c.file), 0o700); e != nil {
		return fmt.Errorf(
			"failed to create directory %s: %w",
			filepath.Dir(c.file),
			e,
		)
	}

	if e = os.WriteFile(c.file, []byte(c.String()), 0o600); e != nil {
		return fmt.Errorf("failed to write %s: %w", c.file, e)
	}

	return nil
}

func (c *config) String() string {
	var b []byte

	b, _ = json.MarshalIndent(&c, "", "  ")
	return strings.TrimSpace(string(b)) + "\n"
}
