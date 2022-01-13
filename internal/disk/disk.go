package disk

import (
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"os/exec"
	"strings"
	"time"
)

// GetDisk returns a block containing the remaining space for the home partition
func GetDisk(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const hddRune = '\uf51f'
	var diskSpace string
	diskBlock := blockutils.Block{
		Name:        blockutils.DiskName,
		Border:      blockutils.Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command("df", "-h")
		outLines, err := blockutils.RunCmdStdout(cmd)
		if err != nil {
			log.FileLog(err)
		}

		if outLines != nil {
			for _, line := range outLines {
				if strings.HasSuffix(line, "/home") {
					diskSpace = strings.Fields(line)[3]
					break
				}
			}
			diskBlock.FullText = string(hddRune) + " " + diskSpace
		}

		blockCh <- &diskBlock
		time.Sleep(timeout)
	}
}
