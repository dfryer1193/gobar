package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Enum for block names
const (
	DiskName    string = "DISK"
	PackName    string = "PACKAGES"
	TempName    string = "TEMPERATURE"
	VolName     string = "VOLUME"
	MediaName   string = "MEDIA"
	DateName    string = "DATE"
	TimeName    string = "TIME"
	BatteryName string = "BATTERY"
)

func runCmdStdout(cmd *exec.Cmd) ([]string, error) {
	var stdoutLines []string
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		fileLog(err)
		return nil, err
	}

	sc := bufio.NewScanner(stdout)

	for sc.Scan() {
		stdoutLines = append(stdoutLines, sc.Text())
	}

	cmd.Wait()

	return stdoutLines, nil
}

func getDisk(timeout time.Duration, blockCh chan<- *block) {
	const hddRune = '\uf7c9'
	var diskSpace string
	diskBlock := block{
		Name:        DiskName,
		Border:      Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command("df", "-h")
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			fileLog(err)
		}

		if outLines != nil {
			for _, line := range outLines {
				if strings.HasSuffix(line, "/home") {
					diskSpace = strings.Fields(line)[3]
					break
				}
			}
			diskBlock.FullText = string(hddRune) + " " + diskSpace
		}

		blockCh <- &diskBlock
		time.Sleep(timeout)
	}
}

func getPackages(timeout time.Duration, blockCh chan<- *block) {
	const updateSym = '\uf560'
	const rebootSym = '\u27f3'
	var prefix string
	homedir, _ := os.UserHomeDir()
	packageCount := 0
	packBlock := block{
		Name:        PackName,
		Border:      Green,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command(homedir + "/.bin/yayupdates")
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			fileLog(err)
		}

		prefix = string(updateSym)
		if outLines != nil {
			packageCount = 0
			for _, line := range outLines {
				packageCount++
				if strings.HasPrefix(line, "linux ") {
					prefix += string(rebootSym)
				}
			}
		}

		packBlock.FullText = prefix + " " + strconv.Itoa(packageCount)

		blockCh <- &packBlock
		time.Sleep(timeout)
	}
}

func getTempFromPath(path string) float64 {
	var tmpBytes []byte
	val := 0.0
	f, err := os.Open(path)
	if err != nil {
		fileLog(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	tmpBytes = sc.Bytes()

	val, err = strconv.ParseFloat(string(tmpBytes), 32)
	if err != nil {
		fileLog(err)
		val = 0.0
	}

	return val / 1000.0
}

func findHWMon() string {
	platformDir := `/sys/devices/platform/`
	coretempRegex := regexp.MustCompile(`coretemp\.[0-9]+`)
	hwmonRegex := regexp.MustCompile(`hwmon[0-9]`)
	platformDevices, err := ioutil.ReadDir(platformDir)
	if err != nil {
		fileLog(err)
		return ""
	}

	for _, pDev := range platformDevices {
		if coretempRegex.MatchString(pDev.Name()) {
			coretempDevices, err := ioutil.ReadDir(platformDir + pDev.Name() + "/hwmon/")
			if err != nil {
				fileLog(err)
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

func getTemp(timeout time.Duration, blockCh chan<- *block) {
	const tempSym = '\uf8c7'
	thermMon := findHWMon()
	if thermMon == "" {
		return
	}
	tempPath := thermMon + "/temp1_input"
	alarmPath := thermMon + "/temp1_crit"
	tempBlock := block{
		Name:        TempName,
		Border:      Blue,
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

func getVolume(timeout time.Duration, blockCh chan<- *block) {
	const soundOnSym = '\uf028'
	const soundOffSym = '\uf026'
	stateRegex := regexp.MustCompile(`\[(on|off)\]`)
	volRegex := regexp.MustCompile(`[0-9]{1,3}%`)
	var state string
	var volume string
	volBlock := block{
		Name:        VolName,
		Border:      White,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}

	for {
		cmd := exec.Command("amixer", "get", "Master")
		outLines, err := runCmdStdout(cmd)
		if err != nil {
			fileLog(err)
		}

		if outLines != nil {
			for _, line := range outLines {
				state = string(stateRegex.Find([]byte(line)))
				volume = string(volRegex.Find([]byte(line)))
				if state == "" || volume == "" {
					continue
				}
			}

			if state == "[off]" {
				volBlock.FullText = string(soundOffSym) + " MUTE"
			} else {
				volBlock.FullText = string(soundOnSym) + " " + volume
			}
		}

		blockCh <- &volBlock

		time.Sleep(timeout)
	}
}

func getPlayers() []string {
	cmd := exec.Command("playerctl", "-l")
	out, err := runCmdStdout(cmd)
	if err != nil {
		fileLog(err)
	}
	return out
}

func getCurPlayer(players []string) string {
	for _, player := range players {
		stateCmd := exec.Command("playerctl", "-p", player, "status")
		state, err := runCmdStdout(stateCmd)
		if err != nil {
			fileLog(err)
		}
		if state[0] == "Playing" {
			return player
		}
	}
	return ""
}

func getMediaState(player string) string {
	status, err := runCmdStdout(exec.Command("playerctl", "-p", player, "status"))
	if err != nil {
		fileLog(err)
	}
	if len(status) > 0 {
		if status[0] == "Playing" {
			return string('\uf04c')
		}
	}

	return string('\uf04b')
}

type mediaState struct {
	lastPlayer string
	text       string
	head       int
	tail       int
}

func (state *mediaState) scroll() string {
	if len(state.text) > 50 {
		if state.head == state.tail {
			state.tail = 46
		}

		if state.tail+1 == len(state.text) {
			state.tail = -1
		}

		if state.head+1 == len(state.text) {
			state.head = -1
		}

		state.tail++
		state.head++

		if state.tail < state.head {
			return state.text[state.head:] + "   " + state.text[:state.tail]
		}

		return state.text[state.head:state.tail] + "..."
	}
	state.head = -1
	state.tail = -1
	return state.text
}

func getMedia(timeout time.Duration, blockCh chan<- *block) {
	fmtStr := `{{ artist }} - {{ title }}`
	mediaBlock := block{
		Name:        MediaName,
		Border:      Red,
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Urgent:      false,
		FullText:    "",
	}
	state := mediaState{
		lastPlayer: "",
		text:       "",
		head:       -1,
		tail:       -1,
	}

	for {
		tmp := getCurPlayer(getPlayers())

		if tmp != "" && tmp != state.lastPlayer {
			state.lastPlayer = tmp
		}

		if state.lastPlayer != "" {
			infoCmd := exec.Command("playerctl", "-p", state.lastPlayer, "metadata", "-f", fmtStr)
			curState, err := runCmdStdout(infoCmd)
			if err != nil {
				fileLog(err)
				continue
			}

			if len(curState) > 0 {
				if curState[0] != "" {
					if curState[0] != state.text {
						state.text = curState[0]
						state.head = -1
						state.tail = -1
					}
					mediaBlock.FullText = getMediaState(state.lastPlayer) + " " + state.scroll()
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
		Name:        DateName,
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
		Name:        TimeName,
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

func getACState(dir string) bool {
	f, err := os.Open(dir + "/online")
	if err != nil {
		fileLog(err)
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		fileLog(err)
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
		fileLog(err)
		return -1
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()

	val, err := strconv.Atoi(sc.Text())
	if err != nil {
		fileLog(err)
		return -1
	}
	return val
}

func showBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fileLog(err)
	}

	if _, err := os.Stat(homedir + "/.i3/.bat_notified"); os.IsNotExist(err) {
		cmd := exec.Command("notify-send", "Low Battery", "Your battery is below 10%")
		cmd.Run()

		f, err := os.Create(homedir + "/.i3/.bat_notified")
		if err != nil {
			fileLog(err)
		}
		f.Close()
	}
}

func clearBatAlert() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fileLog(err)
	}

	if _, err := os.Stat(homedir + "/.i3/.bat_notified"); os.IsNotExist(err) {
		return
	}

	err = os.Remove(homedir + "/.i3/.bat_notified")
	if err != nil {
		fileLog(err)
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
		Name:        BatteryName,
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
			fileLog(err)
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
