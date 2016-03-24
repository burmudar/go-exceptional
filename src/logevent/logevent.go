package logevent

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type LogLevel string

const INFO_LOG_LEVEL LogLevel = "INFO"
const DEBUG_LOG_LEVEL LogLevel = "DEBUG"
const TRACE_LOG_LEVEL LogLevel = "TRACE"
const ERROR_LOG_LEVEL LogLevel = "ERROR"
const EMPTY_LOG_LEVEL LogLevel = ""

type LogEvent struct {
	Timestamp   *time.Time
	Level       LogLevel
	Source      string
	Description string
}

func (e *LogEvent) string() string {
	return fmt.Sprintf("Event: %v | %v | %v | %v", e.Timestamp, e.Level, e.Source, e.Description)
}

func Parse(line string) (*LogEvent, error) {
	event := new(LogEvent)
	line, date := removeFirstBetweenBrackets(line)

	timestamp, err := toTimestamp(date)
	if err != nil {
		return nil, err
	}
	event.Timestamp = timestamp
	line, level := removeLogLevel(line)
	if level == EMPTY_LOG_LEVEL {
		return nil, errors.New("No Log Level found. Log Level cannto be empty")
	}
	event.Level = level

	line, source := removeSource(line)
	if source == "" {
		return nil, errors.New("No Source found. Source cannot be empty")
	}
	event.Source = source
	event.Description = line
	return event, nil
}

func removeSource(line string) (string, string) {
	end := strings.Index(line, "-")
	if end < 0 {
		return line, ""
	}
	source := strings.Trim(line[0:end], " ")
	newLine := line[end+1 : len(line)]
	newLine = strings.TrimLeft(newLine, " ")
	return newLine, source
}

func removeLogLevel(line string) (string, LogLevel) {
	line = strings.TrimLeft(line, " ")
	end := strings.Index(line, " ")
	var level LogLevel = LogLevel(line[0:end])
	newLine := strings.TrimLeft(line[end:len(line)], " ")
	switch level {
	case INFO_LOG_LEVEL:
		return newLine, INFO_LOG_LEVEL
	case DEBUG_LOG_LEVEL:
		return newLine, DEBUG_LOG_LEVEL
	case TRACE_LOG_LEVEL:
		return newLine, TRACE_LOG_LEVEL
	case ERROR_LOG_LEVEL:
		return newLine, ERROR_LOG_LEVEL
	default:
		return line, EMPTY_LOG_LEVEL
	}
}

func toTimestamp(date string) (*time.Time, error) {
	date = strings.Replace(date, ",", ".", 1)
	timestamp, err := time.Parse("2006-01-02 15:04:05.000", date)
	return &timestamp, err
}

func removeFirstBetweenBrackets(line string) (string, string) {
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == end || start > end {
		return line, ""
	}
	return line[end+1 : len(line)], line[start+1 : end]
}
