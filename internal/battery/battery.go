package battery

import (
	"bufio"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

type bat rune

// Enum of battery icons
const (
	BatFull     bat = '\uf240'
	Bat75       bat = BatFull + 1
	Bat50       bat = BatFull + 2
	Bat25       bat = BatFull + 3
	BatAlert    bat = BatFull + 4
	BatCharging bat = '\uf0e7'
)

func getACState(dir string) bool {
	f, err := os.Open(dir + "/online")
	if err != nil {
		log.FileLog(err)
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		log.FileLog(err)
		return false
	}

	if val == 1 {
		return true
	}
	return false
}

func getBatteryPct(dir string) int {
	f, err := os.Open(dir + "/capacity")
	if err != nil {
		log.FileLog(err)
		return -1
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		log.FileLog(err)
		return -1
	}
	return val
}

func showBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.FileLog(err)
	}

	if _, err := os.Stat(homedir + "/.bat_notified"); os.IsNotExist(err) {
		cmd := exec.Command("notify-send", "Low Battery", "Your battery is below 10%")
		cmd.Run()

		f, err := os.Create(homedir + "/.i3/.bat_notified")
		if err != nil {
			log.FileLog(err)
		}
		f.Close()
	}
}

func clearBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.FileLog(err)
	}

	if _, err := os.Stat(homedir + "/.bat_notified"); os.IsNotExist(err) {
		return
	}

	err = os.Remove(homedir + "/.bat_notified")
	if err != nil {
		log.FileLog(err)
	}

}

// GetBattery returns a block containing the system power state
func GetBattery(timeout time.Duration, blockCh chan<- *blockutils.Block) {
	const powerSupplyDir = "/sys/class/power_supply/"
	var charging bool
	batRegex := regexp.MustCompile(`BAT[0-9]+`)
	acRegex := regexp.MustCompile(`AC`)
	batLevel := -1
	hasBattery := false

	batBlock := blockutils.Block{
		Name:        blockutils.BatteryName,
		Border:      blockutils.White,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		powerSupplies, err := ioutil.ReadDir(powerSupplyDir)
		if err != nil {
			log.FileLog(err)
		}
		//TODO: Handle multiple power supplies.
		for _, f := range powerSupplies {
			if acRegex.MatchString(f.Name()) {
				charging = getACState(powerSupplyDir + f.Name())
			}

			if batRegex.MatchString(f.Name()) {
				batLevel = getBatteryPct(powerSupplyDir + f.Name())
				hasBattery = true
			}
		}

		if charging {
			batBlock.FullText = string(BatCharging)
		} else {
			switch {
			case batLevel < 10:
				batBlock.FullText = string(BatAlert)
				showBatAlert()
			case batLevel < 26:
				batBlock.FullText = string(Bat25)
			case batLevel < 51:
				batBlock.FullText = string(Bat50)
			case batLevel < 76:
				batBlock.FullText = string(Bat75)
			default:
				batBlock.FullText = string(BatFull)
			}
		}
		batBlock.FullText += fmt.Sprintf(" %d%%", batLevel)

		if batLevel >= 10 {
			clearBatAlert()
		}

		if hasBattery {
			blockCh <- &batBlock
		} else {
			return
		}

		time.Sleep(timeout)
	}
}
