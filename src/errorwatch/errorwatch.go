package errorwatch

import (
	"errors"
	"fmt"
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

type Event struct {
	Timestamp   *time.Time
	Level       Level
	Source      string
	Description string
}

type causedBy struct {
	Exception string
	Detail    string
}

type ErrorEvent struct {
	Event
	causedBy
}

func (e *Event) string() string {
	return fmt.Sprintf("Event: %v | %v | %v | %v", e.Timestamp, e.Level, e.Source, e.Description)
}

func ParseLogLine(line string) (*Event, error) {
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

	source := matches[3]
	if source == "" {
		return nil, errors.New("No Source found. Source cannot be empty")
	}
	event.Source = source
	event.Description = matches[4]
	return event, nil
}

func ParseCausedBy(line string, e *Event) (*ErrorEvent, error) {
	if e == nil {
		return nil, errors.New("Cannot create ErrorEvent with nil event")
	}
	c, err := extractCausedBy(line)
	if err != nil {
		return nil, err
	}
	errorEvent := newErrorEvent(e, c)

	return errorEvent, nil
}

func newErrorEvent(e *Event, c *causedBy) *ErrorEvent {
	ee := new(ErrorEvent)
	ee.Timestamp = e.Timestamp
	ee.Level = e.Level
	ee.Source = e.Source
	ee.Description = e.Description
	ee.Exception = c.Exception
	ee.Detail = c.Detail
	return ee
}

func (c *causedBy) isEmpty() bool {
	return c.Exception == "" && c.Detail == ""
}

func extractCausedBy(line string) (*causedBy, error) {
	if !CAUSED_BY_REGEX.MatchString(line) {
		return nil, ErrNotCausedByLine
	}
	c := new(causedBy)
	matches := CAUSED_BY_REGEX.FindStringSubmatch(line)
	c.Exception = matches[1]
	if len(matches) > 2 {
		c.Detail = matches[2]
	}
	if c.isEmpty() {
		return nil, errors.New("No exception nor detail extracted from: " + line)
	}
	return c, nil
}

func ContainsCausedBy(line string) bool {
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
