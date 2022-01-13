package media

import (
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"os/exec"
	"time"
)

//TODO: This is lazy - I should find a better way to do this eventually.
var state = mediaState{
	lastPlayer: "",
	text:       "",
	head:       -1,
	tail:       -1,
}

func getPlayers() []string {
	cmd := exec.Command("playerctl", "-l")
	out, err := blockutils.RunCmdStdout(cmd)
	if err != nil {
		log.FileLog(err)
	}
	return out
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

// GetMedia sends media information for the media block
func GetMedia(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	fmtStr := `{{ artist }} - {{ title }}`
	mediaBlock := blockutils.Block{
		Name:        blockutils.MediaName,
		Border:      blockutils.Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		tmp := getCurPlayer(getPlayers())

		if tmp != "" {
			state.lastPlayer = tmp
		}

		if state.lastPlayer != "" {
			infoCmd := exec.Command("playerctl", "-p", state.lastPlayer, "metadata", "-f", fmtStr)
			curState, err := blockutils.RunCmdStdout(infoCmd)
			if err != nil {
				log.FileLog(err)
				continue
			}

			if len(curState) > 0 {
				if curState[0] != "" {
					if curState[0] != state.text {
						state.text = curState[0]
						state.head = -1
						state.tail = -1
					}
					mediaBlock.FullText = getPlayerState(state.lastPlayer) + " " + state.scroll()
				}
			}
		}

		blockCh <- &mediaBlock
		time.Sleep(timeout)
	}
}

/*
// ClickMedia handles click events for the media block.
func ClickMedia(evt *clickutils.Click) {
	action := ""

	switch evt.Button {
	case clickutils.LeftClick:
		action = "play-pause"
	case clickutils.ScrollUp:
		action = "previous"
	case clickutils.ScrollDown:
		action = "next"
	}

	cmd := exec.Command("playerctl", "-p", state.lastPlayer, action)

	if err := cmd.Run(); err != nil {
		log.FileLog("Could not control media:", err)
	}
}*/
