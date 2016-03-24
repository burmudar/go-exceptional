package main

import (
	"bufio"
	"fmt"
	"github.com/hpcloud/tail"
	"logevent"
	"os"
	"strings"
)

type CausedBy struct {
	Exception   string
	Description string
}

func main() {
}

func readLogFileUsingTail() {
	t, _ := tail.TailFile("test.log", tail.Config{Follow: true, ReOpen: true})
	var event *logevent.LogEvent = nil
	for l := range t.Lines {
		line := l.Text
		fmt.Println(line)
		e, err := logevent.Parse(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		if strings.HasPrefix(line, "Caused by:") {
			parts := strings.Split(line, ":")
			var causedBy *CausedBy = new(CausedBy)
			causedBy.Exception = parts[1]
			causedBy.Description = parts[2]
			fmt.Printf("%v was caused by: %v\n", event, causedBy)
		}
	}

}

func readLogFileUsingScanner() {
	file, err := os.Open("simcontrol.log")
	if err != nil {
		fmt.Errorf("Error occured while opening '%v' for reading. Error: %v", "simcontrol.log", err)
	}
	scanner := bufio.NewScanner(file)
	var event *logevent.LogEvent = nil
	for scanner.Scan() {
		line := scanner.Text()
		e, err := logevent.Parse(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		if strings.HasPrefix(line, "Caused by:") {
			parts := strings.Split(line, ":")
			var causedBy *CausedBy = new(CausedBy)
			causedBy.Exception = parts[1]
			causedBy.Description = parts[2]
			fmt.Printf("%v was caused by: %v\n", event, causedBy)
		}
	}
}
