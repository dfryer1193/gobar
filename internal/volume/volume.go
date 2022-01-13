package volume

import (
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"os/exec"
	"regexp"
	"time"
)

// GetVolume returns a block containing the system volume
func GetVolume(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const soundOnSym = '\uf028'
	const soundOffSym = '\uf026'
	stateRegex := regexp.MustCompile(`\[(on|off)\]`)
	volRegex := regexp.MustCompile(`[0-9]{1,3}%`)
	var state string
	var volume string
	volBlock := blockutils.Block{
		Name:        blockutils.VolName,
		Border:      blockutils.White,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command("amixer", "get", "Master")
		outLines, err := blockutils.RunCmdStdout(cmd)
		if err != nil {
			log.FileLog(err)
		}

		if outLines != nil {
			for _, line := range outLines {
				state = string(stateRegex.Find([]byte(line)))
				volume = string(volRegex.Find([]byte(line)))
				if state == "" || volume == "" {
					continue
				}
			}

			if state == "[off]" {
				volBlock.FullText = string(soundOffSym) + " MUTE"
			} else {
				volBlock.FullText = string(soundOnSym) + " " + volume
			}
		}

		blockCh <- &volBlock

		time.Sleep(timeout)
	}
}