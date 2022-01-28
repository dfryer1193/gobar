package disk

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"strconv"
	"syscall"
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
	fsStat := &syscall.Statfs_t{}

	for {
		err := syscall.Statfs("/home", fsStat)
		if err != nil {
			log.FileLog(err)
		}

		// This will cause an overflow eventually...
		free := fsStat.Bfree * uint64(fsStat.Bsize)
		freeGb := free / 1073741824
		freeGbStr := strconv.FormatUint(freeGb, 10)
		d.block.FullText = string(hddRune) + " " + freeGbStr + "G"

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
