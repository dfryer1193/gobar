package main

import (
	"os/exec"
	"time"
)

type mediaState struct {
	lastPlayer string
	text       string
	head       int
	tail       int
}

//TODO: This is lazy - I should find a better way to do this eventually.
var state = mediaState{
	lastPlayer: "",
	text:       "",
	head:       -1,
	tail:       -1,
}

func (state *mediaState) scroll() string {
	if len(state.text) > 50 {
		if state.head == state.tail {
			state.tail = 46
		}

		if state.tail+1 == len(state.text) {
			state.tail = -1
		}

		if state.head+1 == len(state.text) {
			state.head = -1
		}

		state.tail++
		state.head++

		if state.tail < state.head {
			return state.text[state.head:] + "   " + state.text[:state.tail]
		}

		return state.text[state.head:state.tail] + "..."
	}
	state.head = -1
	state.tail = -1
	return state.text
}

func getPlayers() []string {
	cmd := exec.Command("playerctl", "-l")
	out, err := runCmdStdout(cmd)
	if err != nil {
		fileLog(err)
	}
	return out
}

func getCurPlayer(players []string) string {
	for _, player := range players {
		stateCmd := exec.Command("playerctl", "-p", player, "status")
		state, err := runCmdStdout(stateCmd)
		if err != nil {
			fileLog(err)
		}
		if state[0] == "Playing" {
			return player
		}
	}
	return ""
}

func getPlayerState(player string) string {
	status, err := runCmdStdout(exec.Command("playerctl", "-p", player, "status"))
	if err != nil {
		fileLog(err)
	}
	if len(status) > 0 {
		if status[0] == "Playing" {
			return string('\uf04c')
		}
	}

	return string('\uf04b')
}

func getMedia(timeout time.Duration, blockCh chan<- *block) {
	fmtStr := `{{ artist }} - {{ title }}`
	mediaBlock := block{
		Name:        MediaName,
		Border:      Red,
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
			curState, err := runCmdStdout(infoCmd)
			if err != nil {
				fileLog(err)
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

func clickMedia(evt *click) {
	action := ""

	switch evt.Button {
	case leftClick:
		action = "play-pause"
	case scrollUp:
		action = "previous"
	case scrollDown:
		action = "next"
	}

	cmd := exec.Command("playerctl", "-p", state.lastPlayer, action)

	if err := cmd.Run(); err != nil {
		fileLog("Could not control media:", err)
	}
}
