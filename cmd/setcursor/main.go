package main

import (
	"errors"
	"time"

	"github.com/drgrib/mac"
)

func main() {
	timeoutChan := time.After(10 * time.Second)
loop:
	for {
		select {
		case <-timeoutChan:
			panic(errors.New("timed out without Bear activating"))
		default:
			application, err := mac.GetFrontMostApplication()
			if err != nil {
				panic(err)
			}
			if application == "Bear" {
				break loop
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	script := `
	tell application "System Events"
		tell process "Bear"
			key code 126 using {command down}
		end tell
	end tell
	`
	_, err := mac.RunApplescript(script)
	if err != nil {
		panic(err)
	}
}
