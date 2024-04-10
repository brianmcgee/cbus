package cli

import (
	"github.com/charmbracelet/log"
)

var CLI struct {
	Json      bool `name:"json" short:"J" help:"Json output instead of text"`
	Verbosity int  `name:"verbose" short:"v" type:"counter" default:"0" env:"LOG_LEVEL" help:"Set the verbosity of logs e.g. -vv."`

	Agent   Agent   `cmd:"" help:"run the CBus agent"`
	Nkey    NKey    `cmd:"" help:"derive nkey from ed22519 host key"`
	Get     Get     `cmd:"" help:"get a property"`
	Invoke  Invoke  `cmd:"" help:"invoke a method"`
	Keyscan Keyscan `cmd:"" help:"scan ssh hosts and output their NKeys"`
}

func ConfigureLogging() {
	log.SetReportTimestamp(false)

	if CLI.Verbosity == 0 {
		log.SetLevel(log.WarnLevel)
	} else if CLI.Verbosity == 1 {
		log.SetLevel(log.InfoLevel)
	} else if CLI.Verbosity > 1 {
		log.SetLevel(log.DebugLevel)
	}
}
