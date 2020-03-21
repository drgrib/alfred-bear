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
			time.Sleep(250 * time.Millisecond)
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

/*
timeoutChan := time.After(timoutDur)
	i, err := WindowIndex(application, title)
	hasNotOpenYetError := cantGetProcessError(err)
	switch {
	case hasNotOpenYetError:
		isRunning, err := IsRunning(application)
		if err != nil {
			return -1, err
		}
		if isRunning {
			Activate(application)
		}
	case err != nil && !hasNotOpenYetError:
		return i, err
	}

	for i == -1 {
		select {
		case <-timeoutChan:
			err = fmt.Errorf("Failed to find window %#v for %#v within %v",
				title, application, timoutDur)
			return -1, err
		default:
			i, err = WindowIndex(application, title)
			hasNotOpenYetError := cantGetProcessError(err)
			if err != nil && !hasNotOpenYetError {
				return i, err
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
*/
