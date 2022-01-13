package date

import (
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"time"
)

// GetDate returns a block containing the system date information
func GetDate(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const calendarSym = '\uf073'
	dateBlock := blockutils.Block{
		Name:        blockutils.DateName,
		Border:      blockutils.Green,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		now := time.Now()
		dateBlock.FullText = fmt.Sprintf("%s %s %02d.%02d.%02d", string(calendarSym),
			now.Weekday().String()[0:3], now.Month(), now.Day(), now.Year())

		blockCh <- &dateBlock

		time.Sleep(timeout)
	}
}

// ClickDate handles click events for the date block
func ClickDate(evt *clickutils.Click) {
	w := clickutils.GetWidget(evt.Name)
	if w.Cmd == "" {
		w.Cmd = `exec alacritty --hold -t "` + evt.Name + `" -e cal -3`
		w.Width = 525
		w.Height = 170
	}

	switch evt.Button {
	case clickutils.LeftClick:
		err := w.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
