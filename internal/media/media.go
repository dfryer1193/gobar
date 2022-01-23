package media

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"os/exec"
	"time"
)

// Media - a media block
type Media struct {
	block *blockutils.Block
	state *mediaState
}

const name = blockutils.MediaName

func getPlayers() []string {
	cmd := exec.Command("playerctl", "-l")
	out, err := blockutils.RunCmdStdout(cmd)
	if err != nil {
		log.FileLog(err)
	}
	return out
}

// NewMedia - Returns a new media block
func NewMedia() *Media {
	return &Media{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Red,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		state: &mediaState{
			lastPlayer: "",
			text:       "",
			head:       -1,
			tail:       -1,
		},
	}
}

func getCurPlayer(players []string) string {
	for _, player := range players {
		stateCmd := exec.Command("playerctl", "-p", player, "status")
		state, err := blockutils.RunCmdStdout(stateCmd)
		if err != nil {
			log.FileLog(err)
		}
		if state[0] == "Playing" {
			return player
		}
	}
	return ""
}

func getPlayerState(player string) string {
	status, err := blockutils.RunCmdStdout(exec.Command("playerctl", "-p", player, "status"))
	if err != nil {
		log.FileLog(err)
	}
	if len(status) > 0 {
		if status[0] == "Playing" {
			return string('\uf04c')
		}
	}

	return string('\uf04b')
}

// Refresh - Refresh media information for the media block
func (m *Media) Refresh(timeout time.Duration) {
	fmtStr := `{{ artist }} - {{ title }}`

	for {
		tmp := getCurPlayer(getPlayers())

		if tmp != "" {
			m.state.lastPlayer = tmp
		}

		if m.state.lastPlayer != "" {
			infoCmd := exec.Command("playerctl", "-p", m.state.lastPlayer, "metadata", "-f", fmtStr)
			curState, err := blockutils.RunCmdStdout(infoCmd)
			if err != nil {
				log.FileLog(err)
				continue
			}

			if len(curState) > 0 {
				if curState[0] != "" {
					if curState[0] != m.state.text {
						m.state.text = curState[0]
						m.state.head = -1
						m.state.tail = -1
					}
					m.block.FullText = getPlayerState(m.state.lastPlayer) + " " + m.state.scroll()
				}
			}
		}

		time.Sleep(timeout)
	}
}

// String - the string representation of a Media block
func (m *Media) String() string {
	out, err := json.Marshal(m.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}
	return string(out)
}

// Click handles click events for the media block.
func (m *Media) Click(evt *clickutils.Click) {
	action := ""

	switch evt.Button {
	case clickutils.LeftClick:
		action = "play-pause"
	case clickutils.ScrollUp:
		action = "previous"
	case clickutils.ScrollDown:
		action = "next"
	}

	cmd := exec.Command("playerctl", "-p", m.state.lastPlayer, action)

	if err := cmd.Run(); err != nil {
		log.FileLog("Could not control media:", err)
	}
}
