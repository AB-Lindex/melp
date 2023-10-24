package main

import (
	_ "embed"
	"strings"
)

//go:generate bash ./get_version.sh
//go:embed version.txt
var version string

func versionFunc() string {
	return strings.Trim(version, "\n\r")
}
