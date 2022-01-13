package packages

import (
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetPackages returns a block containing the count of system packages to be
// upgraded. It also uses the icon to indicate whether or not the system will
// need a reboot after upgrade
func GetPackages(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const updateSym = '\uf077'
	const rebootSym = '\uf139'
	var prefix string
	homedir, _ := os.UserHomeDir()
	packageCount := 0
	packBlock := blockutils.Block{
		Name:        blockutils.PackName,
		Border:      blockutils.Green,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command(homedir + "/.bin/yayupdates")
		outLines, err := blockutils.RunCmdStdout(cmd)
		if err != nil {
			log.FileLog(err)
		}

		prefix = string(updateSym)
		if outLines != nil {
			packageCount = 0
			for _, line := range outLines {
				packageCount++
				if strings.HasPrefix(line, "linux ") {
					prefix = string(rebootSym)
				}
			}
		}

		packBlock.FullText = prefix + " " + strconv.Itoa(packageCount)

		blockCh <- &packBlock
		time.Sleep(timeout)
	}
}

// ClickPackages handles click events on the package block
func ClickPackages(evt *clickutils.Click) {
	w := clickutils.GetWidget(evt.Name)
	if w.Cmd == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.FileLog("Couldn't get home dir:", err)
		}
		w.Cmd = `exec alacritty --hold -t "` + evt.Name + `" -e ` + homedir + `/.bin/updateNames.sh`
		w.Width = 300
		w.Height = 500
	}
	switch evt.Button {
	case clickutils.LeftClick:
		err := w.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
