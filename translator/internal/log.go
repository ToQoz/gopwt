package internal

import (
	"os"
)

var debugLog = os.Getenv("GOPWT_DEBUG") != ""
