package app

import (
	"flag"
	"os"
	"path/filepath"
)

type Config struct {
	Filename string
}

func (c *Config) Reset() {
	wd, _ := os.Getwd()
	c.Filename = filepath.Join(wd, "openapi.yaml")
}

func (c *Config) Flags(flags *flag.FlagSet) {
	flags.StringVar(&c.Filename, "file", c.Filename, "the path to the openapi file")
}
