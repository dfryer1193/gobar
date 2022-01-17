package volume

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"os/exec"
	"regexp"
	"time"
)

// Volume - a volume block
type Volume struct {
	block *blockutils.Block
}

const name = blockutils.VolName
const soundOnSym = '\uf028'
const soundOffSym = '\uf026'

var stateRegex = regexp.MustCompile(`\[(on|off)\]`)
var volRegex = regexp.MustCompile(`[0-9]{1,3}%`)

// NewVolume - returns a new volume block
func NewVolume() *Volume {
	return &Volume{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.White,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
	}
}

// Refresh - Refreshes the block containing the system volume
func (v *Volume) Refresh(timeout time.Duration) {
	var state string
	var volume string

	for {
		cmd := exec.Command("amixer", "get", "Master")
		outLines, err := blockutils.RunCmdStdout(cmd)
		if err != nil {
			log.FileLog(err)
		}

		if outLines != nil {
			for _, line := range outLines {
				state = string(stateRegex.Find([]byte(line)))
				volume = string(volRegex.Find([]byte(line)))
				if state == "" || volume == "" {
					continue
				}
			}

			if state == "[off]" {
				v.block.FullText = string(soundOffSym) + " MUTE"
			} else {
				v.block.FullText = string(soundOnSym) + " " + volume
			}
		}

		time.Sleep(timeout)
	}
}

// Marshal - Marshals the volume block into json
func (v *Volume) Marshal() []byte {
	out, err := json.Marshal(v.block)
	if err != nil {
		log.FileLog(err)
		return []byte("{}")
	}
	return out
}

// Click - Handles click events for the volume block
func (v *Volume) Click(evt *clickutils.Click) {
	action := ""
	sndChannel := "Master"

	switch evt.Button {
	case clickutils.LeftClick:
		action = "toggle"
	case clickutils.ScrollUp:
		action = "5%+"
	case clickutils.ScrollDown:
		action = "5%-"
	}

	cmd := exec.Command("amixer", "set", sndChannel, action)

	if err := cmd.Run(); err != nil {
		log.FileLog("Could not control volume:", err)
	}
}
