package blockutils

import (
	"bufio"
	"gobar/internal/log"
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

	cmd.Wait()

	return stdoutLines, nil
}
