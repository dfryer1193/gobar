package main

import (
	"encoding/json"
	"fmt"
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
			case "disk":
				disk, err = json.Marshal(blk)
			case "packages":
				pack, err = json.Marshal(blk)
			}
		}
		if err != nil {
			fmt.Printf("Could not marshal JSON!")
			// XXX: log to file
		}

		fmt.Printf(",[%s,%s,%s,%s,%s,%s,%s,%s]", disk,
			pack,
			temp,
			vol,
			media,
			date,
			sysTime,
			bat)
	}
}

//	fmt.Printf(`,[%s%s%s%s%s%s%s%s`,)

func readClicks() {

}

func main() {
	//TODO: Build this around a config file
	setup()
	for {
		fmt.Printf(",[")
		go printBlocks()
		go readClicks()
		fmt.Printf("]")
		time.Sleep(1 * time.Second)
	}
}
