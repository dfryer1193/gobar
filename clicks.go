package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.i3wm.org/i3"
)

type button int

const (
	leftClick button = iota + 1
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

func isShowing(name string) *i3.Node {
	tree, err := i3.GetTree()
	if err != nil {
		fileLog(err)
		return nil
	}

	return tree.Root.FindChild(func(n *i3.Node) bool {
		return n.Name == name
	})
}

func kill(n *i3.Node) {
	i3cmd := fmt.Sprintf(`[con_id=%d] kill`, n.ID)
	i3.RunCommand(i3cmd)
}

func diskWidget(name string, x, y int) error {
	if widget := isShowing(name); widget != nil {
		kill(widget)
		return nil
	}

	sub := i3.Subscribe(i3.WindowEventType)
	defer sub.Close()

	_, err := i3.RunCommand(`exec termite --hold -t "DISK" -e "df -h"`)
	if err != nil {
		return err
	}

	var win i3.Node
	for sub.Next() {
		evt := sub.Event().(*i3.WindowEvent)

		if evt.Change == "new" {
			if evt.Container.Name == name {
				win = evt.Container
				break
			}
		}
	}

	i3cmd := fmt.Sprintf(`[con_id="%d"] move position %d %d`, win.ID, x, y)
	i3.RunCommand(i3cmd)
	return nil
}

func clickDisk(evt *click) {
	switch button(evt.Button) {
	case leftClick:
		diskWidget(evt.Name, evt.X, evt.Y)
	}
}

func handleClicks() {
	var evt click
	rd := bufio.NewReader(os.Stdin)

	for {
		s, err := rd.ReadString('\n')
		if err != nil {
			fileLog(err)
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
			fileLog(err)
		}

		switch evt.Name {
		case "DISK":
			clickDisk(&evt)
		}
	}
}
