package clickutils

import (
	"gobar/internal/log"

	"go.i3wm.org/i3"
)

const (
	// LeftClick - a left click
	LeftClick int = iota + 1
	// MiddleClick - a middle click
	MiddleClick
	// RightClick - a right click
	RightClick
	// ScrollUp - a scroll-up event
	ScrollUp
	// ScrollDown - a scroll-down event
	ScrollDown
)

// Click represents a click event
type Click struct {
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

func getFocusedOutput() *i3.Node {
	tree, err := i3.GetTree()
	if err != nil {
		log.FileLog(err)
	}

	return tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Type == i3.OutputNode
	})
}

// Clickable is an interface that allows things to be clicked
type Clickable interface {
	Click(*Click)
}
