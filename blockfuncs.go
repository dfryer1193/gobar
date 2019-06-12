package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func runCmdStdout(cmd *exec.Cmd) ([]string, error) {
	var stdoutLines []string
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	sc := bufio.NewScanner(stdout)

	for sc.Scan() {
		stdoutLines = append(stdoutLines, sc.Text())
	}

	cmd.Wait()

	return stdoutLines, nil
}

func getDisk(timeout int, blockCh chan<- *block) {
	var diskSpace string
	cmd := exec.Command("df", "-h")
	diskBlock := block{
		Name:        "disk",
		Border:      Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			// XXX: Log err to file
		}

		if outLines != nil {
			for _, line := range outLines {
				if strings.HasSuffix(line, "/home") {
					diskSpace = strings.Split(line, " ")[3]
					break
				}
			}
			diskBlock.FullText = diskSpace
		}

		blockCh <- &diskBlock
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getPackages(timeout int, blockCh chan<- *block) {
	cmd := exec.Command("~/.bin/yayupdates")
	prefix := `⬆`
	packageCount := 0
	packBlock := block{
		Name:        "packages",
		Border:      Green,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			// XXX: Log err to file
		}

		if outLines != nil {
			packageCount = 0
			prefix = `⬆`
			for _, line := range outLines {
				packageCount++
				if strings.HasPrefix(line, "linux ") {
					prefix += `⟳`
				}
			}
		}

		packBlock.FullText = prefix + strconv.Itoa(packageCount)

		blockCh <- &packBlock
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getTempFromPath(path string) float64 {
	var tmpBytes []byte
	val := 0.0
	f, err := os.Open(path)
	if err != nil {
		// XXX: Log error to file
	}
	defer f.Close()

	_, err = f.Read(tmpBytes)
	if err != io.EOF && err != nil {
		// XXX: Log error to file
	}

	val, err = strconv.ParseFloat(string(tmpBytes), 32)
	if err != nil {
		// XXX: Log error to file
		val = 0.0
	}

	return val / 1000.0
}

func getTemp(timeout int, blockCh chan<- *block) {
	tempPath := "/sys/devices/platform/coretemp.0/hwmon/hwmon3/temp1_input"
	alarmPath := "/sys/devices/platform/coretemp.0/hwmon/hwmon3/temp1_crit"
	tempBlock := block{
		Name:        "temperature",
		Border:      Blue,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	thresh := getTempFromPath(alarmPath)

	for {
		tempVal := getTempFromPath(tempPath)
		if tempVal > thresh {
			tempBlock.Urgent = true
		}

		tempBlock.FullText = fmt.Sprintf("+%3.2f°C", tempVal)

		blockCh <- &tempBlock
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getVolume(timeout int, blockCh chan<- *block) {
	stateRegex := regexp.MustCompile(`\[(on|off)\]`)
	volRegex := regexp.MustCompile(`[0-9]{1,3}%`)
	var state string
	var volume string
	cmd := exec.Command("amixer", "get", "Master")
	volBlock := block{
		Name:        "volume",
		Border:      White,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			//XXX: Log err to file
		}

		if outLines != nil {
			for _, line := range outLines {
				state = string(stateRegex.Find([]byte(line)))
				volume = string(volRegex.Find([]byte(line)))
				if state == "" || volume == "" {
					continue
				} else {
					break
				}
			}

			if state == "[off]" {
				volBlock.FullText = ` MUTE`
			} else {
				volBlock.FullText = ` ` + volume
			}
		}

		blockCh <- &volBlock
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getPlayers() ([]string, error) {
	cmd := exec.Command("playerctl", "-l")
	return runCmdStdout(cmd)
}

func getCurPlayer(players []string) string {
	for _, player := range players {
		stateCmd := exec.Command("playerctl", "-p", player, "status")
		state, err := runCmdStdout(stateCmd)
		if err != nil {
			//XXX: log to file
		}
		if state[0] == "Playing" {
			return player
		}
	}
	return ""
}

func getMedia(timeout int, blockCh chan<- *block) {
	var curPlayer string
	fmtStr := `{{ emoji(status) }} {{ artist }} - {{ title }}`
	mediaBlock := block{
		Name:        "media",
		Border:      Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		players, err := getPlayers()
		if err != nil {
			// XXX: Log err to file
		}

		tmp := getCurPlayer(players)

		if tmp != "" && tmp != curPlayer {
			curPlayer = tmp
		}

		if curPlayer != "" {
			infoCmd := exec.Command("playerctl", "-p", curPlayer, "metadata", "-f", fmtStr)
			state, err := runCmdStdout(infoCmd)
			if err != nil {
				// XXX: log to file
				log.Fatal(err)
				//continue
			}

			if len(state) > 0 {
				if state[0] != "" {
					mediaBlock.FullText = state[0]
				}
			}
		}

		blockCh <- &mediaBlock
		time.Sleep(timeout)
	}
}

func getDate(timeout time.Duration, blockCh chan<- *block) {
	const calendarSym = '\uf073'
	dateBlock := block{
		Name:        "date",
		Border:      Green,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		now := time.Now()
		dateBlock.FullText = fmt.Sprintf("%s %s %02d.%02d.%02d", string(calendarSym),
			now.Weekday().String()[0:3], now.Month(), now.Day(), now.Year())

		blockCh <- &dateBlock

		time.Sleep(timeout)
	}
}

func getTime(timeout time.Duration, blockCh chan<- *block) {
	const clockSym = '\uf64f'
	timeBlock := block{
		Name:        "time",
		Border:      Blue,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		now := time.Now()
		timeBlock.FullText = fmt.Sprintf("%s %02d.%02d.%02d", string(clockSym),
			now.Hour(), now.Minute(), now.Second())

		blockCh <- &timeBlock

		time.Sleep(timeout)
	}
}

type bat rune

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

func getACState(dir string) bool {
	f, err := os.Open(dir + "/online")
	if err != nil {
		// XXX: Log err to file
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		//XXX: log err to file
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
		//XXX: Log err to file
		return -1
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		//XXX: Log err to file
		return -1
	}
	return val
}

func showBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		//XXX: Log err to file
	}

	if _, err := os.Stat(homedir + "/.i3/.bat_notified"); os.IsNotExist(err) {
		cmd := exec.Command("notify-send", "Low Battery", "Your battery is below 10%")
		cmd.Run()

		f, err := os.Create(homedir + "/.i3/.bat_notified")
		if err != nil {
			//XXX: Log err to file
		}
		f.Close()
	}
}

func clearBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		//XXX: Log err to file
	}

	if _, err := os.Stat(homedir + "/.i3/.bat_notified"); os.IsNotExist(err) {
		return
	}

	err = os.Remove(homedir + "/.i3/.bat_notified")
	if err != nil {
		//XXX: Log err to file
	}

}

func getBattery(timeout time.Duration, blockCh chan<- *block) {
	const powerSupplyDir = "/sys/class/power_supply/"
	var charging bool
	batRegex := regexp.MustCompile(`BAT[0-9]+`)
	acRegex := regexp.MustCompile(`AC`)
	batLevel := -1
	hasBattery := false

	batBlock := block{
		Name:        "battery",
		Border:      White,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		powerSupplies, err := ioutil.ReadDir(powerSupplyDir)
		if err != nil {
			// XXX: Log err to file
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
			switch {
			case batLevel < 11:
				batBlock.FullText = string(Bat10Charge)
			case batLevel < 26:
				batBlock.FullText = string(Bat25Charge)
			case batLevel < 51:
				batBlock.FullText = string(Bat50Charge)
			case batLevel < 66:
				batBlock.FullText = string(Bat65Charge)
			case batLevel < 81:
				batBlock.FullText = string(Bat80Charge)
			case batLevel < 96:
				batBlock.FullText = string(Bat95Charge)
			default:
				batBlock.FullText = string(BatCharged)
			}
		} else {
			switch {
			case batLevel < 10:
				batBlock.FullText = string(BatAlert)
				showBatAlert()
			case batLevel < 20:
				batBlock.FullText = string(Bat10)
			case batLevel < 30:
				batBlock.FullText = string(Bat20)
			case batLevel < 40:
				batBlock.FullText = string(Bat30)
			case batLevel < 50:
				batBlock.FullText = string(Bat40)
			case batLevel < 60:
				batBlock.FullText = string(Bat50)
			case batLevel < 70:
				batBlock.FullText = string(Bat60)
			case batLevel < 80:
				batBlock.FullText = string(Bat70)
			case batLevel < 90:
				batBlock.FullText = string(Bat80)
			case batLevel < 95:
				batBlock.FullText = string(Bat90)
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
