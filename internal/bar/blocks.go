package bar

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gobar/internal/battery"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/date"
	"gobar/internal/disk"
	"gobar/internal/log"
	"gobar/internal/media"
	"gobar/internal/packages"
	"gobar/internal/systime"
	"gobar/internal/temperature"
	"gobar/internal/volume"
	"os"
	"strings"
	"time"
)

var diskBlk = disk.NewDisk()

// PrintBlocks prints all blocks and handles click events
func PrintBlocks() {
	empty := []byte("{}")

	packJSON := empty
	tempJSON := empty
	volJSON := empty
	mediaJSON := empty
	dateJSON := empty
	sysTimeJSON := empty
	batJSON := empty

	var err error

	blockCh := make(chan *blockutils.Block, 1)

	go diskBlk.Refresh(5 * time.Second)
	go packages.GetPackages(1*time.Hour, blockCh)
	go temperature.GetTemp(1*time.Second, blockCh)
	go volume.GetVolume(1*time.Second, blockCh)
	go media.GetMedia(1*time.Second, blockCh)
	go date.GetDate(1*time.Hour, blockCh)
	go systime.GetTime(1*time.Second, blockCh)
	go battery.GetBattery(5*time.Second, blockCh)

	for {
		select {
		case blk := <-blockCh:
			switch blk.Name {
			case blockutils.PackName:
				packJSON, err = json.Marshal(blk)
			case blockutils.TempName:
				tempJSON, err = json.Marshal(blk)
			case blockutils.VolName:
				volJSON, err = json.Marshal(blk)
			case blockutils.MediaName:
				mediaJSON, err = json.Marshal(blk)
			case blockutils.DateName:
				dateJSON, err = json.Marshal(blk)
			case blockutils.TimeName:
				sysTimeJSON, err = json.Marshal(blk)
			case blockutils.BatteryName:
				batJSON, err = json.Marshal(blk)
			}
		}
		if err != nil {
			log.FileLog(err)
		}

		fmt.Printf(",[%s,%s,%s,%s,%s,%s,%s", diskBlk.Marshal(),
			packJSON,
			tempJSON,
			volJSON,
			mediaJSON,
			dateJSON,
			sysTimeJSON)

		if !bytes.Equal(batJSON, empty) {
			fmt.Printf(",%s", batJSON)
		}

		fmt.Printf("]")
	}
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

		clickers[evt.Name].Click(&evt)

		switch evt.Name {
		case blockutils.PackName:
			packages.ClickPackages(&evt)
		case blockutils.TempName:
			temperature.ClickTemp(&evt)
		case blockutils.VolName:
			volume.ClickVolume(&evt)
		case blockutils.MediaName:
			media.ClickMedia(&evt)
		case blockutils.DateName:
			date.ClickDate(&evt)
		}
	}
}
