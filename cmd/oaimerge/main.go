package main

import (
	"flag"
	"github.com/worldiety/oaimerge/internal/app"
	"os"
)

func main() {
	var cfg app.Config
	cfg.Reset()
	cfg.Flags(flag.CommandLine)
	flag.Parse()

	buf, err := app.Apply(cfg)
	if err != nil {
		panic(err)
	}

	_, _ = os.Stdout.Write(buf)
}
