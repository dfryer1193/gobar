package disk

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"os/exec"
	"strings"
	"time"
)

// Disk represents a disk block
type Disk struct {
	block  *blockutils.Block
	widget *clickutils.Widget
}

const name = blockutils.DiskName
const hddRune = '\uf51f'

// NewDisk - returns a new Disk object
func NewDisk() *Disk {
	return &Disk{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Red,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e df -h`,
			Width:  535,
			Height: 215,
		},
	}
}

// Refresh - refreshes the information for the block on a set interval
func (d *Disk) Refresh(timeout time.Duration) {
	var diskSpace string

	for {
		// TODO: Use syscall.Statfs for this.
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
			d.block.FullText = string(hddRune) + " " + diskSpace
		}

		time.Sleep(timeout)
	}
}

// String - The string representation of a Disk block
func (d *Disk) String() string {
	out, err := json.Marshal(d.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}
	return string(out)
}

// Click - handles click events for the disk block
func (d *Disk) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := d.widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
