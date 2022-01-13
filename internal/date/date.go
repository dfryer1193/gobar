package date

import (
	"fmt"
	"gobar/internal/blockutils"
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
