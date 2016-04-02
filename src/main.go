package main

import (
	"bufio"
	"errors"
	"errorwatch"
	"fmt"
	"github.com/hpcloud/tail"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"net/smtp"
	"os"
	_ "path"
	_ "path/filepath"
	"time"
)

var store errorwatch.Store

var ErrNotCausedByLine error = errors.New("Line does not contain 'Caused by'")

func main() {
	/*
		files := []string{}
		filepath.Walk("logs", func(p string, i os.FileInfo, err error) error {
			if path.Ext(p) == ".log" {
				files = append(files, p)
			}
			return nil
		})
		err := initDB()
		if err != nil {
			fmt.Printf("Failed initializing Database: [%v]\n", err)
			return
		} else {
			fmt.Println("Database initiliazed")
		}
		loadAll(files)
		initStats()
		var eventProcess chan errorwatch.ErrorEvent = make(chan errorwatch.ErrorEvent)
		go processEventsFromChannel(eventProcess)
		watchFile("test.log", eventProcess)
	*/
	store := errorwatch.NewStore()
	store.Init()
	var e *errorwatch.ErrorEvent = new(errorwatch.ErrorEvent)
	t := time.Now()
	e.Timestamp = &t
	e.Level = errorwatch.ERROR_LOG_LEVEL
	e.Source = "test"
	e.Description = "test description"
	e.Exception = "TestException"
	e.Detail = "Some Detail"
	notify(e, nil, nil)
}

func notify(e *errorwatch.ErrorEvent, stat *errorwatch.StatItem, sum *errorwatch.Summary) {

}

func loadAll(files []string) {
	fmt.Printf("Files: %v\n", files)
	for _, filePath := range files {
		fmt.Printf("Loading File: %v\n", filePath)
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			fmt.Errorf("Error occured while opening '%v' for reading. Error: %v", "simcontrol.log", err)
		}
		scanner := bufio.NewScanner(file)
		readLogFile(scanner)
	}
}

func watchFile(path string, errorProcess chan errorwatch.ErrorEvent) {
	t, _ := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
	var event *errorwatch.Event = nil
	fmt.Printf("Tailing file: %v\n", path)
	for l := range t.Lines {
		line := l.Text
		fmt.Printf("Tail: %v\n", line)
		e, err := errorwatch.ParseLogLine(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		errorEvent, err := processEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			errorProcess <- *errorEvent
			event = nil
		}
	}
}

func processEventIfIsCausedByLine(line string, event *errorwatch.Event) (*errorwatch.ErrorEvent, error) {
	if errorwatch.ContainsCausedBy(line) {
		errorEvent, err := errorwatch.ParseCausedBy(line, event)
		if err != nil {
			return nil, err
		} else {
			err = store.Errors().AddErrorEvent(errorEvent)
			if err != nil {
				return nil, err
			} else {
				return errorEvent, nil
			}
		}
	}
	return nil, ErrNotCausedByLine
}
func readLogFile(scanner *bufio.Scanner) {
	var event *errorwatch.Event = nil
	for scanner.Scan() {
		line := scanner.Text()
		e, err := errorwatch.ParseLogLine(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		errorEvent, err := processEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			event = nil
		}
	}
}
