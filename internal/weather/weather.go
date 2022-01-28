package weather

import (
	"encoding/json"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"io"
	"net/http"
	"strings"
	"time"
)

// Weather - a weather block
type Weather struct {
	block  *blockutils.Block
	widget *clickutils.Widget
}

const name = blockutils.WeatherName
const weatherURL = `https://wttr.in/`
const formatStr = `%c%t`
const queryParams = `?format=` + formatStr

// NewWeather returns a new weather block
func NewWeather() *Weather {
	return &Weather{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Blue,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e curl v2.wttr.in`,
			Width:  592,
			Height: 766,
		},
	}
}

// Refresh - refreshes the weather widget
func (w *Weather) Refresh(timeout time.Duration) {
	for {
		resp, err := http.Get(weatherURL + queryParams)

		if err != nil {
			log.FileLog(err)
			w.block.FullText = ""
			time.Sleep(timeout / 10)
			continue
		}

		if resp.StatusCode != 200 {
			log.FileLog(resp.Status)
			w.block.FullText = ""
			time.Sleep(timeout / 10)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.FileLog("Cannot read wttr.in response body")
			w.block.FullText = ""
			continue
		}

		if strings.HasPrefix(string(body), "Unknown location") {
			log.FileLog("wttr.in is down")
			w.block.FullText = ""
			continue

		}

		w.block.FullText = string(body)

		time.Sleep(timeout)
	}
}

// String - the string representation of a Weather block
func (w *Weather) String() string {
	out, err := json.Marshal(w.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}

	return string(out)
}

// Click handles click events for the temperature block
func (w *Weather) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := w.widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
