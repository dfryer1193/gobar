package bar

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gobar/internal/battery"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/date"
	"gobar/internal/disk"
	"gobar/internal/log"
	"gobar/internal/media"
	"gobar/internal/network"
	"gobar/internal/packages"
	"gobar/internal/systime"
	"gobar/internal/temperature"
	"gobar/internal/volume"
	"gobar/internal/weather"
	"os"
	"strings"
	"time"
)

var weatherBlk = weather.NewWeather()
var diskBlk = disk.NewDisk()
var packBlk = packages.NewPackages()
var tempBlk = temperature.NewTemperature()
var volBlk = volume.NewVolume()
var mediaBlk = media.NewMedia()
var dateBlk = date.NewDate()
var timeBlk = systime.NewSystime()
var batBlk = battery.NewBattery()
var netBlk = network.NewNetwork()

// PrintBlocks prints all blocks and handles click events
func PrintBlocks() {
	hasBattery := battery.HasBattery()

	go weatherBlk.Refresh(1 * time.Hour)
	go diskBlk.Refresh(5 * time.Second)
	go packBlk.Refresh(1 * time.Hour)
	go tempBlk.Refresh(1 * time.Second)
	go volBlk.Refresh(1 * time.Second)
	go mediaBlk.Refresh(1 * time.Second)
	go dateBlk.Refresh(1 * time.Hour)
	go timeBlk.Refresh(1 * time.Second)
	go netBlk.Refresh(30 * time.Second)
	if hasBattery {
		go batBlk.Refresh(5 * time.Second)
	}

	for {
		fmt.Printf(
			",[%s, %s,%s,%s,%s,%s,%s,%s, %s",
			weatherBlk,
			diskBlk,
			packBlk,
			tempBlk,
			volBlk,
			mediaBlk,
			dateBlk,
			timeBlk,
			netBlk,
		)

		if hasBattery {
			fmt.Printf(",%s", batBlk)
		}

		fmt.Printf("]")
		time.Sleep(1 * time.Second)
	}
}

var clickers = map[string]clickutils.Clickable{
	blockutils.WeatherName: weatherBlk,
	blockutils.DiskName:    diskBlk,
	blockutils.PackName:    packBlk,
	blockutils.TempName:    tempBlk,
	blockutils.VolName:     volBlk,
	blockutils.MediaName:   mediaBlk,
	blockutils.DateName:    dateBlk,
	blockutils.NetworkName: netBlk,
}

// HandleClicks handles click events for the bar
func HandleClicks() {
	var evt clickutils.Click
	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			log.FileLog("Input err", err)
			continue
		}
		s := sc.Text()

		if strings.HasPrefix(s, ",") {
			s = s[1:]
		}

		if strings.HasPrefix(s, "[") {
			continue
		}

		if err := json.Unmarshal([]byte(s), &evt); err != nil {
			log.FileLog("JSON Unmarshal err", err)
		}

		if clicker, ok := clickers[evt.Name]; ok {
			clicker.Click(&evt)
		}
	}
}
