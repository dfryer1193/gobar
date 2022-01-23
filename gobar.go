package main

import (
	"fmt"
	"sync"

	"gobar/internal/bar"
)

func setup() {
	fmt.Printf(`{ "version": 1, "click_events": true }[[]`)
}

func main() {
	//TODO: Build this around a config file
	wg := sync.WaitGroup{}
	wg.Add(2)
	setup()
	go bar.PrintBlocks()
	go bar.HandleClicks()
	wg.Wait()
}
