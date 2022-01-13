package systime

import (
	"fmt"
	"gobar/internal/blockutils"
	"time"
)

// GetTime returns a block containing the system time
func GetTime(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const clockSym = '\uf017'
	timeBlock := blockutils.Block{
		Name:        blockutils.TimeName,
		Border:      blockutils.Blue,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		now := time.Now()
		timeBlock.FullText = fmt.Sprintf("%s %02d.%02d.%02d", string(clockSym),
			now.Hour(), now.Minute(), now.Second())

		blockCh <- &timeBlock

		time.Sleep(timeout)
	}
}
