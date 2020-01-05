package logging

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

func init() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
}

// Init Sets the logging levels
func Init(verbose bool) {
	minLevel := logutils.LogLevel("WARN")
	if verbose {
		minLevel = logutils.LogLevel("DEBUG")
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: minLevel,
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
}
