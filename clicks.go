package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdirkse/i3ipc"
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

func getWinIDsByName(sock *i3ipc.IPCSocket, name string) []int64 {
	var ids []int64
	root, err := sock.GetTree()
	if err != nil {
		logErr(err)
		return nil
	}
	wins := root.FindNamed(name)

	for _, win := range wins {
		ids = append(ids, win.ID)
	}

	return ids
}

func toggleWidget(sock *i3ipc.IPCSocket, name string, x, y int) error {
	diskWins := getWinIDsByName(sock, name)
	if len(diskWins) > 0 {
		for _, winID := range diskWins {
			strID := strconv.FormatInt(winID, 10)
			_, err := sock.Command(`[con_id=` + strID + `] kill`)
			if err != nil {
				return err
			}
		}
		return nil
	}

	xStr := strconv.Itoa(x)
	yStr := strconv.Itoa(y)

	switch name {
	case "DISK":
		return diskWidget(sock, name, xStr, yStr)
	}

	return nil
}

func diskWidget(sock *i3ipc.IPCSocket, name, x, y string) error {
	sock.Command(`exec termite --hold -t "DISK" -e "df -h"`)
	time.Sleep(100 * time.Millisecond) // wait for the window to be created
	sock.Command(`[title="^DISK$"] move position ` + x + ` ` + y)
	return nil
}

func clickDisk(sock *i3ipc.IPCSocket, evt *click) {
	switch button(evt.Button) {
	case leftClick:
		toggleWidget(sock, evt.Name, evt.X, evt.Y)
	}
}

func clickPackages(sock *i3ipc.IPCSocket, evt *click) {
	switch button(evt.Button) {
	case leftClick:
		toggleWidget(sock, evt.Name, evt.X, evt.Y)
	}
}

func handleClicks() {
	var evt click
	rd := bufio.NewReader(os.Stdin)

	for {
		s, err := rd.ReadString('\n')
		if err != nil {
			logErr(err)
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
			logErr(err)
		}

		sock, err := i3ipc.GetIPCSocket()
		if err != nil {
			logErr(err)
			return
		}
		switch evt.Name {
		case "DISK":
			clickDisk(sock, &evt)
		}
	}
}
