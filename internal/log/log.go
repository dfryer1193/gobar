package log

import (
	"log"
	"os"
)

const logFile string = "/.i3/gobar.log"

// FileLog logs its args to a file in the local filesystem
func FileLog(v ...interface{}) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	f, err := os.OpenFile(homedir+logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	logger := log.New(f, "", log.Ldate|log.Ltime)

	logger.Println(v...)
}
