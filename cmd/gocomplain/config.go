package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the gocomplain config structure.
type Config struct {
	Confidence float64  `json:"confidence"`
	Ignore     []string `json:"ignore"`
	Length     uint     `json:"length"`
	Over       uint     `json:"over"`
	Prune      []string `json:"prune"`
	Quiet      bool     `json:"quiet"`
	Skip       []string `json:"skip"`
}

var config *Config

func init() {
	var b []byte
	var e error
	var fn string

	if fn, e = os.UserConfigDir(); e != nil {
		panic(fmt.Errorf("user has no config directory: %w", e))
	}

	fn = filepath.Join(fn, "gocomplain", "rc")
	b, e = os.ReadFile(fn)

	if (e != nil) || (len(bytes.TrimSpace(b)) == 0) {
		// Default config
		config = &Config{
			Confidence: 0.8,
			Ignore:     []string{},
			Length:     70,
			Over:       15,
			Prune:      []string{},
			Skip:       []string{},
		}

		b, _ = json.MarshalIndent(&config, "", "  ")

		if e = os.MkdirAll(filepath.Dir(fn), 0o700); e != nil {
			e = fmt.Errorf(
				"failed to create directory %s: %w",
				filepath.Dir(fn),
				e,
			)
			panic(e)
		}

		if e = os.WriteFile(fn, append(b, '\n'), 0o600); e != nil {
			panic(fmt.Errorf("failed to write %s: %w", fn, e))
		}
	} else {
		if e = json.Unmarshal(b, &config); e != nil {
			panic(fmt.Errorf("invalid config: %w", e))
		}
	}

	if config.Confidence == 0 {
		config.Confidence = 0.8
	}

	if config.Ignore == nil {
		config.Ignore = []string{}
	}

	if config.Length == 0 {
		config.Length = 70
	}

	if config.Over == 0 {
		config.Over = 15
	}

	if config.Prune == nil {
		config.Prune = []string{}
	}

	if config.Skip == nil {
		config.Skip = []string{}
	}
}
