package main

import (
	"log"
	"os"
)

const logFile string = "/.i3/gobar.log"

func fileLog(v ...interface{}) {
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
