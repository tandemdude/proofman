package logging

import (
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"strings"
)

func log(message string, args ...any) {
	fmt.Printf(strings.TrimSpace(message)+"\n", args...)
}

func Verbose(message string, args ...any) {
	if internal.LogLevel >= internal.LogLvlVerbose {
		log(message, args...)
	}
}

func Unquiet(message string, args ...any) {
	if internal.LogLevel >= internal.LogLvlUnquiet {
		log(message, args...)
	}
}

func Quiet(message string, args ...any) {
	if internal.LogLevel >= internal.LogLvlQuiet {
		log(message, args...)
	}
}
