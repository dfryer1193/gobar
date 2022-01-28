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
	BatFull     bat = '\uf578'
	Bat10       bat = BatFull + 1
	Bat20       bat = BatFull + 2
	Bat30       bat = BatFull + 3
	Bat40       bat = BatFull + 4
	Bat50       bat = BatFull + 5
	Bat60       bat = BatFull + 6
	Bat70       bat = BatFull + 7
	Bat80       bat = BatFull + 8
	Bat90       bat = BatFull + 9
	BatAlert    bat = BatFull + 10
	BatCharged  bat = '\uf584'
	Bat10Charge bat = BatCharged + 1
	Bat25Charge bat = BatCharged + 2
	Bat50Charge bat = BatCharged + 3
	Bat65Charge bat = BatCharged + 4
	Bat80Charge bat = BatCharged + 5
	Bat95Charge bat = BatCharged + 6
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
			switch {
			case batLevel < 11:
				b.block.FullText = string(Bat10Charge)
			case batLevel < 26:
				b.block.FullText = string(Bat25Charge)
			case batLevel < 51:
				b.block.FullText = string(Bat50Charge)
			case batLevel < 66:
				b.block.FullText = string(Bat65Charge)
			case batLevel < 81:
				b.block.FullText = string(Bat80Charge)
			case batLevel < 96:
				b.block.FullText = string(Bat95Charge)
			default:
				b.block.FullText = string(BatCharged)
			}
		} else {
			switch {
			case batLevel < 10:
				b.block.FullText = string(BatAlert)
				showBatAlert()
			case batLevel < 20:
				b.block.FullText = string(Bat10)
			case batLevel < 30:
				b.block.FullText = string(Bat20)
			case batLevel < 40:
				b.block.FullText = string(Bat30)
			case batLevel < 50:
				b.block.FullText = string(Bat40)
			case batLevel < 60:
				b.block.FullText = string(Bat50)
			case batLevel < 70:
				b.block.FullText = string(Bat60)
			case batLevel < 80:
				b.block.FullText = string(Bat70)
			case batLevel < 90:
				b.block.FullText = string(Bat80)
			case batLevel < 95:
				b.block.FullText = string(Bat90)
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

// String - Gets the string representation of the battery block
func (b *Battery) String() string {
	out, err := json.Marshal(b.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}
	return string(out)
}
