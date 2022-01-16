package temperature

import (
	"bufio"
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

var coretempRegex = regexp.MustCompile(`coretemp\.[0-9]+`)
var hwmonRegex = regexp.MustCompile(`hwmon[0-9]`)

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

// GetTemp returns a block containing the cpu temperature
func GetTemp(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const tempSym = '\uf769'
	thermMon := findHWMon()
	if thermMon == "" {
		return
	}
	tempPath := thermMon + "/temp1_input"
	alarmPath := thermMon + "/temp1_crit"
	tempBlock := blockutils.Block{
		Name:        blockutils.TempName,
		Border:      blockutils.Blue,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	thresh := getTempFromPath(alarmPath)

	for {
		tempVal := getTempFromPath(tempPath)
		if tempVal > thresh {
			tempBlock.Urgent = true
		}

		tempBlock.FullText = fmt.Sprintf("%s %3.1f°C", string(tempSym), tempVal)

		blockCh <- &tempBlock
		time.Sleep(timeout)
	}
}

// ClickTemp handles click events for the temperature block
func ClickTemp(evt *clickutils.Click) {
	w := clickutils.GetWidget(evt.Name)
	if w.Cmd == "" {
		w.Cmd = `exec alacritty --hold -t "` + evt.Name + `" -e echo $(cat /proc/loadavg | cut -d \  -f -3)`
		w.Width = 115
		w.Height = 50
	}

	switch evt.Button {
	case clickutils.LeftClick:
		err := w.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
