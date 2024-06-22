package main

import "github.com/mjwhitta/jsoncfg"

var config *jsoncfg.JSONCfg

func init() {
	// Initialize default values for config
	config = jsoncfg.New("~/.config/gocomplain/rc")
	_ = config.SetDefault(0.8, "confidence")
	_ = config.SetDefault([]string{}, "ignore")
	_ = config.SetDefault(70, "length")
	_ = config.SetDefault(15, "over")
	_ = config.SetDefault([]string{}, "prune")
	_ = config.SetDefault(false, "quiet")
	_ = config.SetDefault([]string{}, "skip")
	_ = config.SaveDefault()
	_ = config.Reset()
}
