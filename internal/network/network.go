package network

import (
	"encoding/json"
	"fmt"
	"gobar/internal/blockutils"
	"gobar/internal/clickutils"
	"gobar/internal/log"
	"time"

	gnm "github.com/Wifx/gonetworkmanager"
)

// Network - a network block
type Network struct {
	block  *blockutils.Block
	widget *clickutils.Widget
}

const name = blockutils.NetworkName
const disconnected = '\ufaa9'
const connectedWifi = '\ufaa8'
const connectedLan = '\uf817'

// NewNetwork returns a new network block
func NewNetwork() *Network {
	return &Network{
		block: &blockutils.Block{
			Name:        name,
			Border:      blockutils.Green,
			BorderLeft:  0,
			BorderRight: 0,
			BorderTop:   0,
			Urgent:      false,
			FullText:    "",
		},
		widget: &clickutils.Widget{
			Title:  name,
			Cmd:    `exec alacritty --hold -t "` + name + `" -e nmtui`,
			Width:  676,
			Height: 530,
		},
	}
}

func getState(nm gnm.NetworkManager) gnm.NmState {
	state, err := nm.State()
	if err != nil {
		log.FileLog(err)
		return gnm.NmStateUnknown
	}

	return state
}

// Refresh - refreshes the network widget
func (n *Network) Refresh(timeout time.Duration) {
	nm, err := gnm.NewNetworkManager()
	if err != nil {
		log.FileLog("Could not connect to NetworkManager")
		return
	}

	for {
		state := getState(nm)
		if state == gnm.NmStateUnknown {
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		if state != gnm.NmStateConnectedGlobal {
			n.block.FullText = string(disconnected) + " "
			time.Sleep(timeout)
			continue
		}

		conn, err := nm.GetPropertyPrimaryConnection()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		devices, err := conn.GetPropertyDevices()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		if len(devices) < 1 {
			log.FileLog("Primary connection has no devices")
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		dev := devices[0]
		devType, err := dev.GetPropertyDeviceType()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		if devType != gnm.NmDeviceTypeWifi {
			n.block.FullText = string(connectedLan) + " "
			time.Sleep(timeout)
			continue
		}

		ap, err := conn.GetPropertySpecificObject()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		ssid, err := ap.GetPropertySSID()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		strength, err := ap.GetPropertyStrength()
		if err != nil {
			log.FileLog(err)
			n.block.FullText = ""
			time.Sleep(timeout)
			continue
		}

		n.block.FullText = fmt.Sprintf(
			"%s %d%% %s",
			string(connectedWifi),
			strength,
			ssid,
		)

		time.Sleep(timeout)
	}
}

// String - the string representation of a Network block
func (n *Network) String() string {
	out, err := json.Marshal(n.block)
	if err != nil {
		log.FileLog(err)
		return "{}"
	}

	return string(out)
}

// Click - handles click events for the click block
func (n *Network) Click(evt *clickutils.Click) {
	switch evt.Button {
	case clickutils.LeftClick:
		err := n.widget.Toggle(evt.X, evt.Y)
		if err != nil {
			log.FileLog(err)
		}
	}
}
