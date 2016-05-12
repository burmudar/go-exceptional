package errord

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hpcloud/tail"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Level string

const INFO_LOG_LEVEL Level = "INFO"
const DEBUG_LOG_LEVEL Level = "DEBUG"
const TRACE_LOG_LEVEL Level = "TRACE"
const ERROR_LOG_LEVEL Level = "ERROR"
const EMPTY_LOG_LEVEL Level = ""
const DATE_FORMAT string = "2006-01-02 15:04:05.000"

const CAUSED_BY string = "Caused by:"

var ErrNotLogLine error = errors.New("Line does not match Log Line format")
var ErrNotCausedByLine error = errors.New("Line does not match Caused by format or does not contain 'Caused by'")

var LOG_LINE_REGEX = regexp.MustCompile(`^\[([\w\d\s-:,]+)\]\s(INFO|ERROR|TRACE|DEBUG)\s+([\w\d.:]+)\s-\s(.*)`)

var CAUSED_BY_REGEX = regexp.MustCompile(`Caused by:\s([\w\d\.]+):?\s?(.*)`)

type ErrorParser interface {
	Parse(src string) ParseStats
	Watch(src string) chan ErrorEvent
}

type LogFileParser struct {
	errorStorage  ErrorStore
	metricStorage MetricStore
}

type Event struct {
	Timestamp   *time.Time
	Level       Level
	Description string
}

type ErrorEvent struct {
	Event
	Exception string
	Detail    string
}

type MetricEvent struct {
	Event
	Name  string
	Value float32
}

type ParseStats struct {
	Lines   int
	Failed  int
	Success int
}

func (p ParseStats) string() string {
	return fmt.Sprintf("Lines [%v] Failed [%v] Succeeded[%v]", p.Lines, p.Failed, p.Success)
}

func (e *ErrorEvent) string() string {
	return fmt.Sprintf("Event: %v | %v | %v", e.Timestamp, e.Level, e.Description)
}

func (e *ErrorEvent) hasCausedBy() bool {
	if e.Exception != "" {
		return true
	}
	return false
}

func NewLogFileParser(errorStorage ErrorStore, metricStorage MetricStore) ErrorParser {
	return &LogFileParser{errorStorage, metricStorage}
}

func (p *LogFileParser) Parse(src string) ParseStats {
	var stats ParseStats
	file, err := os.Open(src)
	defer file.Close()
	if err != nil {
		log.Printf("Error occured while opening '%v' for reading. Error: %v", src, err)
		return stats
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		stats.Lines++
		event, err := parseLogLine(line)
		if err != nil {
			stats.Failed++
		}
		errorEvent, err := createErrorEvent(line, event)
		if err != nil {
			continue
		}
		err = p.errorStorage.Add(errorEvent)
		if err != nil {
			log.Printf("Failed inserting Event[%v - %v]", event.Timestamp, errorEvent.Exception)
		} else {
			stats.Success++
		}
	}
	return stats
}

func (p *LogFileParser) Watch(src string) chan ErrorEvent {
	//Should add some way to stop go routine. Maybe errorStorage the Tail t variable since it might have a stop method ?
	eventBus := make(chan ErrorEvent)
	go func() {
		t, _ := tail.TailFile(src, tail.Config{Follow: true, ReOpen: true})
		for l := range t.Lines {
			line := l.Text
			event, err := parseLogLine(line)
			errorEvent, err := createErrorEvent(line, event)
			if err != nil {
				continue
			}
			err = p.errorStorage.Add(errorEvent)
			if err != nil {
				log.Printf("Failed inserting Event[%v - %v] -> %v", errorEvent.Timestamp, errorEvent.Exception, err)
			}
			log.Printf("Passing Event to ErrorChan!")
			eventBus <- *errorEvent
		}
	}()
	return eventBus
}

func createErrorEvent(line string, event *Event) (*ErrorEvent, error) {
	if event == nil {
		return nil, errors.New("Cannot create ErrorEvent with nil event")
	}
	if containsCausedBy(line) {
		errorEvent := &ErrorEvent{Event: *event}
		excp, detail, err := parseCausedBy(line)
		errorEvent.Exception = excp
		errorEvent.Detail = detail
		if err != nil {
			return nil, err
		} else if !errorEvent.hasCausedBy() {
			return nil, errors.New("No exception nor detail extracted from: " + line)
		} else {
			return errorEvent, nil
		}
	}
	return nil, ErrNotCausedByLine
}

func parseLogLine(line string) (*Event, error) {
	if !LOG_LINE_REGEX.MatchString(line) {
		return nil, ErrNotLogLine
	}
	event := new(Event)

	matches := LOG_LINE_REGEX.FindStringSubmatch(line)

	timestamp, err := toTimestamp(matches[1])
	if err != nil {
		return nil, err
	}
	event.Timestamp = timestamp
	level := Level(matches[2])
	if level == EMPTY_LOG_LEVEL {
		return nil, errors.New("No Log Level found. Log Level cannto be empty")
	}
	event.Level = level

	//Skipping matches[3] which is supposed to be the source
	event.Description = matches[4]
	return event, nil
}

func parseCausedBy(line string) (excp, detail string, err error) {
	excp = ""
	detail = ""
	if !CAUSED_BY_REGEX.MatchString(line) {
		return excp, detail, ErrNotCausedByLine
	}
	matches := CAUSED_BY_REGEX.FindStringSubmatch(line)
	excp = matches[1]
	if len(matches) >= 2 {
		detail = matches[2]
	}
	return excp, detail, nil
}

func containsCausedBy(line string) bool {
	line = strings.TrimLeft(line, " ")
	if strings.HasPrefix(line, CAUSED_BY) {
		return true
	}
	return false
}

func toTimestamp(date string) (*time.Time, error) {
	date = strings.Replace(date, ",", ".", 1)
	timestamp, err := time.Parse(DATE_FORMAT, date)
	return &timestamp, err
}
