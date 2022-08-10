package main

import (
	"github.com/ninlil/butler/log"
)

func printErrors(errs []error, format string, v ...interface{}) bool {
	if len(errs) == 0 {
		return false
	}

	if len(errs) == 1 {
		v = append(v, errs[0])
		log.Error().Msgf(format+" %v", v...)
		return true
	}

	log.Error().Msgf(format, v...)
	for _, err := range errs {
		log.Error().Msgf("  -> %v", err)
	}

	return true
}
