package testutils

import (
	"flag"
)

var dFlag = flag.Bool("debug", false, "enable debug mode")

func init() {
}

func IsDebug() bool {
	return *dFlag
}
