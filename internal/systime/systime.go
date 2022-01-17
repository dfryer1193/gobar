package systime

import (
	"encoding/json"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"time"
)

const name = blockutils.TimeName

// Systime - a block showing the system time
type Systime struct {
	block *blockutils.Block
}

// NewSystime - Returns a new system time block
func NewSystime() *Systime {
	return &Systime{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Blue,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
	}
}

// Refresh - Refreshes the block containing the system time
func (s *Systime) Refresh(timeout time.Duration) {
	const clockSym = '\uf017'

	for {
		now := time.Now()
		s.block.FullText = fmt.Sprintf("%s %02d.%02d.%02d", string(clockSym),
			now.Hour(), now.Minute(), now.Second())

		time.Sleep(timeout)
	}
}

// Marshal - Marshals the time block into json
func (s *Systime) Marshal() []byte {
	out, err := json.Marshal(s.block)
	if err != nil {
		log.FileLog(err)
		return []byte("{}")
	}
	return out
}
