package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Color - Create a type to store our color strings.
type Color string

// Create an enum of colors
const (
	Black   Color = `#222222`
	Red     Color = `#e84f4f`
	Green   Color = `#b7ce42`
	Blue    Color = `#66aabb`
	Magenta Color = `#b7416e`
	Cyan    Color = `#6d878d`
	White   Color = `#dddddd`
	None    Color = ``
)

type block struct {
	Name        string `json:"name"`
	Border      Color  `json:"border"`
	BorderLeft  int    `json:"border_left"`
	BorderRight int    `json:"border_right"`
	BorderTop   int    `json:"border_top"`
	Urgent      bool   `json:"urgent"`
	FullText    string `json:"full_text"`
}

func setup() {
	fmt.Printf(`{ "version": 1, "click_events": true }[[]`)
}

func printBlocks() {
	disk := []byte("{}")
	pack := []byte("{}")
	temp := []byte("{}")
	vol := []byte("{}")
	media := []byte("{}")
	date := []byte("{}")
	sysTime := []byte("{}")
	bat := []byte("{}")
	var err error

	blockCh := make(chan *block, 7)
	go getDisk(5*time.Second, blockCh)
	go getPackages(1*time.Hour, blockCh)
	go getTemp(1*time.Second, blockCh)
	go getVolume(1*time.Second, blockCh)
	go getMedia(5*time.Second, blockCh)
	go getDate(1*time.Hour, blockCh)
	go getTime(1*time.Second, blockCh)
	go getBattery(5*time.Second, blockCh)

	for {
		select {
		case blk := <-blockCh:
			switch blk.Name {
			case "DISK":
				disk, err = json.Marshal(blk)
			case "PACKAGES":
				pack, err = json.Marshal(blk)
			case "TEMP":
				temp, err = json.Marshal(blk)
			case "VOLUME":
				vol, err = json.Marshal(blk)
			case "MEDIA":
				media, err = json.Marshal(blk)
			case "DATE":
				date, err = json.Marshal(blk)
			case "TIME":
				sysTime, err = json.Marshal(blk)
			case "BATTERY":
				bat, err = json.Marshal(blk)
			}
		}
		if err != nil {
			fileLog(err)
		}

		fmt.Printf(",[%s,%s,%s,%s,%s,%s,%s", disk,
			pack,
			temp,
			vol,
			media,
			date,
			sysTime)

		if !bytes.Equal(bat, []byte("{}")) {
			fmt.Printf(",%s", bat)
		}

		fmt.Printf("]")
	}
}

func main() {
	//TODO: Build this around a config file
	wg := sync.WaitGroup{}
	wg.Add(2)
	setup()
	go printBlocks()
	go handleClicks()
	wg.Wait()
}
