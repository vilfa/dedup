package log

import (
	"fmt"
	"log"
	"os"
)

func Default() *log.Logger {
	hostname, err := os.Hostname()
	if err != nil {
		log.Panicf("could not get hostname: %v", err)
	}

	return log.New(
		os.Stderr, fmt.Sprintf("[%v/%v/%v] ", hostname, "dedup", os.Getpid()), log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
}
