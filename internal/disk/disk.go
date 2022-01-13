package disk

import (
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
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

// ClickDisk - handles click events for the disk block
func ClickDisk(evt *clickutils.Click) {
	w := clickutils.GetWidget(evt.Name)
	if w.Cmd == "" {
		w.Cmd = `exec alacritty --hold -t "` + evt.Name + `" -e df -h`
		w.Width = 535
		w.Height = 215
	}

	switch evt.Button {
	case clickutils.LeftClick:
		err := w.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
