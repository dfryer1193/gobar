package battery

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/log"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type bat rune

// Battery - A block displaying the battery state
type Battery struct {
	block *blockutils.Block
}

const powerSupplyDir = "/sys/class/power_supply/"
const name = blockutils.BatteryName

// Enum of battery icons
const (
	BatFull     bat = '\uf240'
	Bat75       bat = BatFull + 1
	Bat50       bat = BatFull + 2
	Bat25       bat = BatFull + 3
	BatAlert    bat = BatFull + 4
	BatCharging bat = '\uf0e7'
)

var batRegex = regexp.MustCompile(`BAT[0-9]+`)
var acRegex = regexp.MustCompile(`AC`)

// NewBattery - Returns a new battery block
func NewBattery() *Battery {
	return &Battery{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.White,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
	}
}

// HasBattery - Returns true if a battery is found for the system
func HasBattery() bool {
	powerSupplies, err := ioutil.ReadDir(powerSupplyDir)
	if err != nil {
		log.FileLog(err)
		return false
	}

	for _, f := range powerSupplies {
		if batRegex.MatchString(f.Name()) {
			return true
		}
	}

	return false
}

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
	text, err := os.ReadFile(dir + "/capacity")
	if err != nil {
		log.FileLog(err)
		return -1
	}

	pct := strings.ReplaceAll(string(text), "\n", "")
	val, err := strconv.Atoi(pct)
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

		f, err := os.Create(homedir + "/.bat_notified")
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

// Refresh - Refreshes the block containing the system power state
func (b *Battery) Refresh(timeout time.Duration) {
	var charging bool
	batLevel := -1
	powerSupplies, err := ioutil.ReadDir(powerSupplyDir)
	if err != nil {
		log.FileLog(err)
	}

	for {
		//TODO: Handle multiple power supplies.
		for _, f := range powerSupplies {
			if acRegex.MatchString(f.Name()) {
				charging = getACState(powerSupplyDir + f.Name())
			}

			if batRegex.MatchString(f.Name()) {
				batLevel = getBatteryPct(powerSupplyDir + f.Name())
			}
		}

		if charging {
			b.block.FullText = string(BatCharging)
		} else {
			switch {
			case batLevel < 10:
				b.block.FullText = string(BatAlert)
				showBatAlert()
			case batLevel < 26:
				b.block.FullText = string(Bat25)
			case batLevel < 51:
				b.block.FullText = string(Bat50)
			case batLevel < 76:
				b.block.FullText = string(Bat75)
			default:
				b.block.FullText = string(BatFull)
			}
		}
		b.block.FullText += fmt.Sprintf(" %d%%", batLevel)

		if batLevel >= 10 {
			clearBatAlert()
		}

		time.Sleep(timeout)
	}
}

// Marshal - Marshals the battery block into json
func (b *Battery) Marshal() []byte {
	out, err := json.Marshal(b.block)
	if err != nil {
		log.FileLog(err)
		return []byte("{}")
	}
	return out
}
