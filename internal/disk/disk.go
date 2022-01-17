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
	Block  *blockutils.Block
	Widget *clickutils.Widget
}

// NewDisk - returns a new Disk object
func NewDisk() *Disk {
	name := blockutils.DiskName
	return &Disk{
		Block: &blockutils.Block{
			Name:        blockutils.DiskName,
			Border:      blockutils.Red,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		Widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e df -h`,
			Width:  535,
			Height: 215,
		},
	}
}

// Refresh - refreshes the information for the block on a set interval
func (d *Disk) Refresh(timeout time.Duration) {
	const hddRune = '\uf51f'
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
			d.Block.FullText = string(hddRune) + " " + diskSpace
		}

		time.Sleep(timeout)
	}
}

// Marshal - Marshals the disk block into json
func (d *Disk) Marshal() []byte {
	out, err := json.Marshal(d.Block)
	if err != nil {
		log.FileLog(err)
		return []byte("{}")
	}
	return out
}

// Click - handles click events for the disk block
func (d *Disk) Click(evt *clickutils.Click) {
	w := d.Widget
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
