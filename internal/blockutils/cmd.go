package blockutils

import (
	"bufio"
	"gobar/internal/log"
	"os"
	"os/exec"
)

// RunCmdStdout runs a command, then captures and returns the output
func RunCmdStdout(cmd *exec.Cmd) ([]string, error) {
	var stdoutLines []string
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.FileLog(err)
		return nil, err
	}

	sc := bufio.NewScanner(stdout)

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			log.FileLog(err)
			return nil, err
		}
		stdoutLines = append(stdoutLines, sc.Text())
	}

	if err := cmd.Wait(); err != nil {
		log.FileLog(err)
		return nil, err
	}
	return stdoutLines, nil
}

// Homedir - returns the user's homedir
func Homedir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.FileLog("Couldn't get homedir: ", err)
	}

	return homedir
}
