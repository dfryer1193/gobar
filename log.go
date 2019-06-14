package main

import (
	"log"
	"os"
)

const logFile string = "/.i3/gobar.log"

func logErr(e error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}

	f, err := os.OpenFile(home+logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write([]byte(e.Error() + "\n")); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func logStr(s string) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}

	f, err := os.OpenFile(home+logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write([]byte(s + "\n")); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
