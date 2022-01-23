package packages

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Packages represents a package block
type Packages struct {
	Block  *blockutils.Block
	Widget *clickutils.Widget
}

const name = blockutils.PackName

var homedir = blockutils.Homedir()

// NewPackages - returns a new Packages object
func NewPackages() *Packages {
	return &Packages{
		Block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Green,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		Widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e ` + homedir + `/.bin/updateNames.sh`,
			Width:  300,
			Height: 500,
		},
	}
}

// Refresh updates the Packages block with the count of system packages to be
// upgraded. It also uses the icon to indicate whether or not the system will
// need a reboot after upgrade
func (p *Packages) Refresh(timeout time.Duration) {
	const updateSym = '\uf077'
	const rebootSym = '\uf139'
	var prefix string
	packageCount := 0

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

		p.Block.FullText = prefix + " " + strconv.Itoa(packageCount)

		time.Sleep(timeout)
	}
}

//String - the string representation of a Media block
func (p *Packages) String() string {
	out, err := json.Marshal(p.Block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}

	return string(out)
}

// Click - handles click events on the package block
func (p *Packages) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := p.Widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
