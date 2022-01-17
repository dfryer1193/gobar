package date

import (
	"encoding/json"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"time"
)

// Date - A date block
type Date struct {
	block  *blockutils.Block
	widget *clickutils.Widget
}

const name = blockutils.DateName
const calendarSym = '\uf073'

// NewDate - returns a new date block
func NewDate() *Date {
	return &Date{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Green,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e cal -3`,
			Width:  525,
			Height: 170,
		},
	}
}

// Refresh - Refreshes the block containing the system date information
func (d *Date) Refresh(timeout time.Duration) {
	for {
		now := time.Now()
		d.block.FullText = fmt.Sprintf("%s %s %02d.%02d.%02d", string(calendarSym),
			now.Weekday().String()[0:3], now.Month(), now.Day(), now.Year())

		time.Sleep(timeout)
	}
}

// Marshal - Marshals the date block into json
func (d *Date) Marshal() []byte {
	out, err := json.Marshal(d.block)
	if err != nil {
		log.FileLog(err)
		return []byte("{}")
	}
	return out
}

// Click - Handles click events for the date block
func (d *Date) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := d.widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
