package main

import "github.com/mjwhitta/jsoncfg"

var config *jsoncfg.JSONCfg

func init() {
	// Initialize default values for config
	config = jsoncfg.New("~/.config/gocomplain/rc")
	config.SetDefault(0.8, "confidence")
	config.SetDefault([]string{}, "ignore")
	config.SetDefault(70, "length")
	config.SetDefault(15, "over")
	config.SetDefault([]string{}, "prune")
	config.SetDefault(false, "quiet")
	config.SetDefault([]string{}, "skip")
	config.SaveDefault()
	config.Reset()
}
