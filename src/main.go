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

func processEventsFromChannel(eventChan chan errorwatch.ErrorEvent) {
	var start time.Time = time.Now()
	var statCache map[string]*errorwatch.StatItem = make(map[string]*errorwatch.StatItem)
	for event := range eventChan {
		if isEventAfterStart(&start, &event) {
			start = time.Now()
			fmt.Println("Event is day after we started. Purging Stat cache")
			statCache = make(map[string]*errorwatch.StatItem)
			fmt.Println("Recalculating stats")
			calcStats()
		}
		fmt.Printf("Processing: %v\n", event)
		var statItem *errorwatch.StatItem
		var ok bool
		if statItem, ok = statCache[event.Exception]; !ok {
			statItem = store.Stats().GetStatItem(event.Exception)
			statCache[event.Exception] = statItem
		}
		if statItem == nil {
			notify(&event, nil, nil)
		} else {
			fmt.Printf("Stats: %v\n", *statItem)
			var s *errorwatch.Summary = store.Stats().GetDaySummary(&event)
			fmt.Printf("errorwatch.Summary: %v\n", *s)
			max := statItem.Mean + statItem.StdDev
			fmt.Printf("Total[%v] : Max[%v]\n", s.Total, max)
			if s.Total >= int(max) {
				notify(&event, statItem, s)
			} else {
			}
		}

	}
}

func isEventAfterStart(start *time.Time, event *errorwatch.ErrorEvent) bool {
	fmt.Printf("%v : %v", start.Day(), event.Timestamp.Day())
	return start.Day()-event.Timestamp.Day() > 0
}

func notify(e *errorwatch.ErrorEvent, stat *errorwatch.StatItem, sum *errorwatch.Summary) {
	if store.Notifications().HasNotification(e) {
		fmt.Printf("Notification already sent for %v\n", *e)
	}
	from := "USERNAME"
	pass := "PASSWORD"
	to := "TO SOMEONE"
	subject := ""
	body := ""
	if stat == nil && sum == nil {
		subject = "Subject: " + fmt.Sprintf("New Error: %v", e.Exception)
		body = fmt.Sprintf("New Error Event: [%v] - [%v] : [%v]\nCaused by: [%v] - [%v]\n", e.Timestamp, e.Source, e.Description, e.Exception, e.Detail)
	} else {
		subject = "Subject: " + fmt.Sprintf("Error exceeds Statistical Limit: %v", e.Exception)
		body = fmt.Sprintf("Error Event: [%v] - [%v] : [%v]\nCaused by: [%v] - [%v]\n Seen today = %v\n Max = %v", e.Timestamp, e.Source, e.Description, e.Exception, e.Detail, sum.Total, stat.Mean+stat.StdDev)
	}

	msg := fmt.Sprintf("From: %v\nTo: %v\nSubject: %v\n\n%v", from, to, subject, body)
	if err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"), from, []string{to}, []byte(msg)); err != nil {
		fmt.Printf("Failed sending notification for [%v] : %v\n", *e, err)
	} else {
		store.Notifications().UpdateNotificationSent(e)
	}

}

func initStats() {
	err := store.Stats().UpdateDaySummaries()
	if err != nil {
		fmt.Printf("Failed loading Day Summaries: [%v]\n", err)
	}
	fmt.Println("Day summaries for errors initialized")

	calcStats()
}
func calcStats() {
	summaries := store.Stats().FetchDaySummaries()
	var statMap map[string][]errorwatch.Summary = make(map[string][]errorwatch.Summary)
	for _, s := range summaries {
		if item, ok := statMap[s.Exception]; ok {
			statMap[s.Exception] = append(item, s)
		} else {
			statMap[s.Exception] = append([]errorwatch.Summary{}, s)
		}
	}
	for k, v := range statMap {
		total := calcTotal(v)
		avg := float64(total / len(v))
		variance := calcVariance(v, avg)
		stdDev := math.Sqrt(float64(variance))
		now := time.Now()
		statItem := errorwatch.StatItem{k, avg, variance, stdDev, total, len(v), &now}
		err := store.Stats().InsertOrUpdateStatItem(&statItem)
		if err != nil {
			fmt.Printf("Failed inserting Stat Item for: [%v] : %v\n", k, err)
		} else {
			fmt.Printf("Inserted errorwatch.StatItem for -> %v\n", k)
		}
	}
}

func calcTotal(summaries []errorwatch.Summary) int {
	var total int
	for _, s := range summaries {
		total += s.Total
	}
	return total
}

func calcVariance(summaries []errorwatch.Summary, avg float64) int {
	var variance int
	for _, s := range summaries {
		diff := float64(s.Total) - avg
		variance += int(math.Pow(diff, 2))
	}
	return variance / len(summaries)
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
