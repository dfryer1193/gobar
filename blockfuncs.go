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

		curPlayer = getCurPlayer(players)

		if curPlayer != "" {
			infoCmd := exec.Command("playerctl", "-p", "-f", fmtStr)
			state, err := runCmdStdout(infoCmd)
			if err != nil {
				// XXX: log to file
			}
			if state[0] != "" {
				mediaBlock.FullText = state[0]
			}
		}

		blockCh <- &mediaBlock
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getDate(timeout int, blockCh chan<- *block) {
	dateBlock := block{
		Name:        "date",
		Border:      "Green",
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		dateBlock.FullText = time.Now().Format("Fri 06.07.19")

		blockCh <- &dateBlock

		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getTime(timeout int, blockCh chan<- *block) {
	timeBlock := block{
		Name:        "date",
		Border:      "Green",
		BorderLeft:  0,
		BorderRight: 0,
		BorderTop:   0,
		Background:  None,
		Urgent:      false,
		FullText:    "",
	}

	for {
		timeBlock.FullText = time.Now().Format("13.00.00")

		blockCh <- &timeBlock

		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func getBattery(timeout int, blockCh chan<- *block) {

}
