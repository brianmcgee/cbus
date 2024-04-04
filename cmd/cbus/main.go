package main

import (
	"github.com/alecthomas/kong"
	"github.com/brianmcgee/cbus/internal/cmd/cli"
)

func main() {
	ctx := kong.Parse(&cli.CLI)
	cli.ConfigureLogging()
	ctx.FatalIfErrorf(ctx.Run())
}
