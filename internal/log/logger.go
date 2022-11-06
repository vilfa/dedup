package log

import (
	"fmt"
	"log"
	"os"
)

func Default() *log.Logger {
	return log.New(
		os.Stderr, fmt.Sprintf("dedup[%d]: ", os.Getpid()), log.LUTC)
}
