package bar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gobar/internal/battery"
	"gobar/internal/blockutils"
	"gobar/internal/date"
	"gobar/internal/disk"
	"gobar/internal/log"
	"gobar/internal/media"
	"gobar/internal/packages"
	"gobar/internal/systime"
	"gobar/internal/temperature"
	"gobar/internal/volume"
	"time"
)

// PrintBlocks prints all blocks and handles click events
func PrintBlocks() {
	diskJSON := []byte("{}")
	packJSON := []byte("{}")
	tempJSON := []byte("{}")
	volJSON := []byte("{}")
	mediaJSON := []byte("{}")
	dateJSON := []byte("{}")
	sysTimeJSON := []byte("{}")
	batJSON := []byte("{}")
	var err error

	blockCh := make(chan *blockutils.Block, 7)
	go disk.GetDisk(5*time.Second, blockCh)
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
			case blockutils.DiskName:
				diskJSON, err = json.Marshal(blk)
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

		fmt.Printf(",[%s,%s,%s,%s,%s,%s,%s", diskJSON,
			packJSON,
			tempJSON,
			volJSON,
			mediaJSON,
			dateJSON,
			sysTimeJSON)

		if !bytes.Equal(batJSON, []byte("{}")) {
			fmt.Printf(",%s", batJSON)
		}

		fmt.Printf("]")
	}
}
