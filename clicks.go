package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.i3wm.org/i3"
)

const (
	leftClick int = iota + 1
	middleClick
	rightClick
	scrollUp
	scrollDown
)

type click struct {
	Name      string   `json:"name"`
	Instance  string   `json:"instance"`
	Button    int      `json:"button"`
	Modifiers []string `json:"modifiers"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	RelativeX int      `json:"relative_x"`
	RelativeY int      `json:"relative_y"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
}

type widget struct {
	title  string
	cmd    string
	node   *i3.Node
	width  int64
	height int64
}

var widgetMap = make(map[string]*widget)

func getFocusedOutput() *i3.Node {
	tree, err := i3.GetTree()
	if err != nil {
		fileLog(err)
	}

	return tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Type == i3.OutputNode
	})
}

func (n *widget) setPosition(x, y int) (int, int) {
	adjX, adjY := int64(x), int64(y)
	output := getFocusedOutput()

	if adjX+n.width > (output.Rect.Width + output.Rect.X) {
		adjX = (output.Rect.Width - n.width) + output.Rect.X
	}

	if adjY+n.height > (output.Rect.Height + output.Rect.Y) {
		adjY = (output.Rect.Height - n.height) + output.Rect.Y
	}

	return int(adjX), int(adjY)
}

func (n *widget) kill() {
	i3cmd := fmt.Sprintf(`[con_id=%d] kill`, n.node.ID)
	i3.RunCommand(i3cmd)
	n.node = nil
}

func (n *widget) toggle(x, y int) error {
	if n.node != nil {
		n.kill()
		return nil
	}

	sub := i3.Subscribe(i3.WindowEventType)
	defer sub.Close()

	_, err := i3.RunCommand(n.cmd)
	if err != nil {
		return err
	}

	for sub.Next() {
		evt := sub.Event().(*i3.WindowEvent)
		if evt.Change == "new" {
			if evt.Container.Name == n.title {
				n.node = &evt.Container
				break
			}
		}
	}

	x, y = n.setPosition(x, y)
	resizeCmd := fmt.Sprintf(`[con_id="%d"] resize set %d %d`, n.node.ID, n.width, n.height)
	moveCmd := fmt.Sprintf(`[con_id="%d"] move position %d %d`, n.node.ID, x, y)
	i3.RunCommand(resizeCmd)
	i3.RunCommand(moveCmd)
	return nil
}

func getWidget(name string) *widget {
	if widgetMap[name] == nil {
		widgetMap[name] = &widget{
			title:  name,
			cmd:    "",
			node:   nil,
			width:  100,
			height: 100,
		}
	}

	return widgetMap[name]
}

func clickDisk(evt *click) {
	w := getWidget(evt.Name)
	if w.cmd == "" {
		w.cmd = `exec termite --hold -t "` + evt.Name + `" -e "df -h"`
		w.width = 535
		w.height = 215
	}

	switch evt.Button {
	case leftClick:
		err := w.toggle(evt.X, evt.Y)
		if err != nil {
			fileLog(err)
		}
	}
}

func clickPackages(evt *click) {
	w := getWidget(evt.Name)
	if w.cmd == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fileLog("Couldn't get home dir:", err)
		}
		w.cmd = `exec termite --hold -t "` + evt.Name + `" -e "` + homedir + `/.bin/updateNames.sh"`
		w.width = 300
		w.height = 500
	}
	switch evt.Button {
	case leftClick:
		err := w.toggle(evt.X, evt.Y)
		if err != nil {
			fileLog(err)
		}
	}
}

func clickTemp(evt *click) {
	w := getWidget(evt.Name)
	if w.cmd == "" {
		w.cmd = `exec termite --hold -t "` + evt.Name + `" -e "echo $(cat /proc/loadavg | cut -d \  -f -3)"`
		w.width = 115
		w.height = 50
	}

	switch evt.Button {
	case leftClick:
		err := w.toggle(evt.X, evt.Y)
		if err != nil {
			fileLog(err)
		}
	}
}

func clickVolume(evt *click) {
	action := ""
	sndChannel := "Master"

	switch evt.Button {
	case leftClick:
		action = "toggle"
	case scrollUp:
		action = "5%+"
	case scrollDown:
		action = "5%-"
	}

	cmd := exec.Command("amixer", "set", sndChannel, action)

	if err := cmd.Run(); err != nil {
		fileLog("Could not control volume:", err)
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

	cmd := exec.Command("playerctl", action)

	if err := cmd.Run(); err != nil {
		fileLog("Could not control media:", err)
	}
}

func clickDate(evt *click) {
	w := getWidget(evt.Name)
	if w.cmd == "" {
		w.cmd = `exec termite --hold -t "` + evt.Name + `" -e "cal -3"`
		w.width = 525
		w.height = 170
	}

	switch evt.Button {
	case leftClick:
		err := w.toggle(evt.X, evt.Y)
		if err != nil {
			fileLog(err)
		}
	}
}

func handleClicks() {
	var evt click
	rd := bufio.NewReader(os.Stdin)

	for {
		s, err := rd.ReadString('\n')
		if err != nil {
			fileLog("Input err", err)
			continue
		}

		if strings.HasPrefix(s, ",") {
			s = s[1:]
		}

		if strings.HasPrefix(s, "[") {
			continue
		}

		err = json.Unmarshal([]byte(s), &evt)
		if err != nil {
			fileLog("JSON Unmarshal err", err)
		}

		switch evt.Name {
		case DISK_NAME:
			clickDisk(&evt)
		case PACK_NAME:
			clickPackages(&evt)
		case TEMP_NAME:
			clickTemp(&evt)
		case VOL_NAME:
			clickVolume(&evt)
		case MEDIA_NAME:
			clickMedia(&evt)
		case DATE_NAME:
			clickDate(&evt)
		}
	}
}
