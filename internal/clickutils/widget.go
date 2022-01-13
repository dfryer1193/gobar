package clickutils

import (
	"fmt"

	"go.i3wm.org/i3"
)

// Widget represents a widget to be toggled on click
type Widget struct {
	Title  string
	Cmd    string
	Node   *i3.Node
	Width  int64
	Height int64
}

func (n *Widget) setPosition(x, y int) (int, int) {
	adjX, adjY := int64(x), int64(y)
	output := getFocusedOutput()

	if adjX+n.Width > (output.Rect.Width + output.Rect.X) {
		adjX = (output.Rect.Width - n.Width) + output.Rect.X
	}

	if adjY+n.Height > (output.Rect.Height + output.Rect.Y) {
		adjY = (output.Rect.Height - n.Height) + output.Rect.Y
	}

	return int(adjX), int(adjY)
}

func (n *Widget) kill() {
	i3cmd := fmt.Sprintf(`[con_id=%d] kill`, n.Node.ID)
	i3.RunCommand(i3cmd)
	n.Node = nil
}

// Toggle - toggles widget state (displayed, destroyed)
func (n *Widget) Toggle(x, y int) error {
	if n.Node != nil {
		n.kill()
		return nil
	}

	sub := i3.Subscribe(i3.WindowEventType)
	defer sub.Close()

	_, err := i3.RunCommand(n.Cmd)
	if err != nil {
		return err
	}

	for sub.Next() {
		evt := sub.Event().(*i3.WindowEvent)
		if evt.Change == "new" {
			if evt.Container.Name == n.Title {
				n.Node = &evt.Container
				break
			}
		}
	}

	x, y = n.setPosition(x, y)
	resizeCmd := fmt.Sprintf(`[con_id="%d"] resize set %d %d`, n.Node.ID, n.Width, n.Height)
	moveCmd := fmt.Sprintf(`[con_id="%d"] move position %d %d`, n.Node.ID, x, y)
	i3.RunCommand(resizeCmd)
	i3.RunCommand(moveCmd)
	return nil
}

var widgetMap = make(map[string]*Widget)

// GetWidget returns a widget with the given name
func GetWidget(name string) *Widget {
	if widgetMap[name] == nil {
		widgetMap[name] = &Widget{
			Title:  name,
			Cmd:    "",
			Node:   nil,
			Width:  100,
			Height: 100,
		}
	}

	return widgetMap[name]
}
