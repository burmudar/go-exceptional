package errorwatch

import (
	"bufio"
	"errors"
	"fmt"
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
	Parse(src *os.File) ParseStats
}

type LogFileParser struct {
	store ErrorStore
}

type TailFileParser struct {
	store ErrorStore
}

type ErrorEvent struct {
	Exception   string
	Detail      string
	Timestamp   *time.Time
	Level       Level
	Source      string
	Description string
}

type ParseStats struct {
	Lines   int
	Failed  int
	Success int
}

func (p *ParseStats) string() string {
	return fmt.Sprintf("Lines [%v] Failed [%v] Succeeded[%v]", p.Lines, p.Failed, p.Success)
}

func (e *ErrorEvent) string() string {
	return fmt.Sprintf("Event: %v | %v | %v | %v", e.Timestamp, e.Level, e.Source, e.Description)
}

func (e *ErrorEvent) hasCausedBy() bool {
	if e.Exception != "" {
		return true
	}
	return false
}

func (p *LogFileParser) Parse(src *os.File) ParseStats {
	scanner := bufio.NewScanner(src)
	var event *ErrorEvent = nil
	var stats ParseStats
	for scanner.Scan() {
		line := scanner.Text()
		stats.Lines++
		e, err := parseLogLine(line)
		if err != nil {
			stats.Failed++
		} else {
			event = e
		}
		err = addCausedBy(line, event)
		if err != nil {
			continue
		}
		err = p.store.Add(event)
		if err != nil {
			log.Printf("Failed inserting Event[%v - %v]", event.Timestamp, event.Exception)
		}
	}
	return stats
}

func addCausedBy(line string, event *ErrorEvent) error {
	if containsCausedBy(line) {
		err := parseCausedBy(line, event)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
	return ErrNotCausedByLine
}

func parseLogLine(line string) (*ErrorEvent, error) {
	if !LOG_LINE_REGEX.MatchString(line) {
		return nil, ErrNotLogLine
	}
	event := new(ErrorEvent)

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

	source := matches[3]
	if source == "" {
		return nil, errors.New("No Source found. Source cannot be empty")
	}
	event.Source = source
	event.Description = matches[4]
	return event, nil
}

func parseCausedBy(line string, e *ErrorEvent) error {
	if e == nil {
		return errors.New("Cannot create ErrorEvent with nil event")
	}
	if !CAUSED_BY_REGEX.MatchString(line) {
		return ErrNotCausedByLine
	}
	matches := CAUSED_BY_REGEX.FindStringSubmatch(line)
	e.Exception = matches[1]
	e.Detail = ""
	if len(matches) > 2 {
		e.Detail = matches[2]
	}
	if e.hasCausedBy() {
		return errors.New("No exception nor detail extracted from: " + line)
	}
	return nil
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
