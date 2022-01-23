package temperature

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"
)

// Temperature - a block for displaying the temperature
type Temperature struct {
	block  *blockutils.Block
	widget *clickutils.Widget
}

const name = blockutils.TempName
const tempSym = '\uf769'

var coretempRegex = regexp.MustCompile(`coretemp\.[0-9]+`)
var hwmonRegex = regexp.MustCompile(`hwmon[0-9]`)
var thermMon = findHWMon()
var tempPath = thermMon + "/temp1_input"
var alarmPath = thermMon + "/temp1_crit"
var thresh = getTempFromPath(alarmPath)

// NewTemperature - returns a new temperature block
func NewTemperature() *Temperature {
	return &Temperature{
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
			Cmd:    `exec alacritty --hold -t "` + name + `" -e htop`,
			Width:  664,
			Height: 168,
		},
	}
}

func getTempFromPath(path string) float64 {
	var tmpBytes []byte
	val := 0.0
	f, err := os.Open(path)
	if err != nil {
		log.FileLog(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	tmpBytes = sc.Bytes()

	val, err = strconv.ParseFloat(string(tmpBytes), 32)
	if err != nil {
		log.FileLog(err)
		val = 0.0
	}

	return val / 1000.0
}

func findHWMon() string {
	platformDir := `/sys/devices/platform/`
	platformDevices, err := ioutil.ReadDir(platformDir)
	if err != nil {
		log.FileLog(err)
		return ""
	}

	for _, pDev := range platformDevices {
		if coretempRegex.MatchString(pDev.Name()) {
			coretempDevices, err := ioutil.ReadDir(platformDir + pDev.Name() + "/hwmon/")
			if err != nil {
				log.FileLog(err)
				return ""
			}
			for _, dev := range coretempDevices {
				if hwmonRegex.MatchString(dev.Name()) {
					return platformDir + pDev.Name() + "/hwmon/" + dev.Name()
				}
			}
		}
	}
	return ""
}

// Refresh - refreshes the block containing the cpu temperature
func (t *Temperature) Refresh(timeout time.Duration) {
	for {
		tempVal := getTempFromPath(tempPath)
		if tempVal > thresh {
			t.block.Urgent = true
		} else {
			t.block.Urgent = false
		}

		t.block.FullText = fmt.Sprintf("%s %3.1f°C", string(tempSym), tempVal)

		time.Sleep(timeout)
	}
}

// String - the string representation of a Temperature block
func (t *Temperature) String() string {
	out, err := json.Marshal(t.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}
	return string(out)
}

// Click handles click events for the temperature block
func (t *Temperature) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := t.widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
