package logevent

import (
	"testing"
	"time"
)

func TestParseOfINFOLine(t *testing.T) {
	const INFO_LINE string = "[2016-03-23 15:41:48,564] INFO  worker.DealerBalanceUpdater:27 - Starting update of dealer balance"
	logEvent, err := Parse(INFO_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When LogEvent is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid INFO log line should not return an error: [%v]", err)
	}

	if logEvent.Level != INFO_LOG_LEVEL {
		t.Errorf("LogEvent should have INFO LogLevel. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 564*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("LogEvent does not contain the Timestamp as its defined in the INFO line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.DealerBalanceUpdater:27"
	if logEvent.Source != expctedSource {
		t.Errorf("LogEvent does not contain source as its defined in the INFO line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "Starting update of dealer balance"
	if logEvent.Description != expectedDescription {
		t.Errorf("LogEvent does not contain description as its defined in the INFO line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseOfTRACELine(t *testing.T) {
	const TRACE_LINE string = "[2016-03-23 15:41:48,608] TRACE worker.SimConsumerImpl:183 - null : Retrieving Balance of Sim id: 0 from Recharge Service"
	logEvent, err := Parse(TRACE_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When LogEvent is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid TRACE log line should not return an error: [%v]", err)
	}

	if logEvent.Level != TRACE_LOG_LEVEL {
		t.Errorf("LogEvent should have TRACE LogLevel. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 608*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("LogEvent does not contain the Timestamp as its defined in the TRACE line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.SimConsumerImpl:183"
	if logEvent.Source != expctedSource {
		t.Errorf("LogEvent does not contain source as its defined in the TRACE line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "null : Retrieving Balance of Sim id: 0 from Recharge Service"
	if logEvent.Description != expectedDescription {
		t.Errorf("LogEvent does not contain description as its defined in the TRACE line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseOfDEBUGLine(t *testing.T) {
	const DEBUG_LINE string = "[2016-03-23 15:41:48,615] DEBUG worker.SimConsumerImpl:129 - null : Sending recharge to service"
	logEvent, err := Parse(DEBUG_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When LogEvent is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid DEBUG log line should not return an error: [%v]", err)
	}

	if logEvent.Level != DEBUG_LOG_LEVEL {
		t.Errorf("LogEvent should have DEBUG LogLevel. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 615*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("LogEvent does not contain the Timestamp as its defined in the DEBUG line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.SimConsumerImpl:129"
	if logEvent.Source != expctedSource {
		t.Errorf("LogEvent does not contain source as its defined in the DEBUG line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "null : Sending recharge to service"
	if logEvent.Description != expectedDescription {
		t.Errorf("LogEvent does not contain description as its defined in the DEBUG line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseOfERRORLine(t *testing.T) {
	const ERROR_LINE string = "[2016-03-23 15:41:48,939] ERROR client.AirtelService:54 - 0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	logEvent, err := Parse(ERROR_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When LogEvent is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid ERROR log line should not return an error: [%v]", err)
	}

	if logEvent.Level != ERROR_LOG_LEVEL {
		t.Errorf("LogEvent should have ERROR LogLevel. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 939*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("LogEvent does not contain the Timestamp as its defined in the ERROR line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "client.AirtelService:54"
	if logEvent.Source != expctedSource {
		t.Errorf("LogEvent does not contain source as its defined in the ERROR line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	if logEvent.Description != expectedDescription {
		t.Errorf("LogEvent does not contain description as its defined in the ERROR line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestRemoveSource(t *testing.T) {
	var expectedSource string = "worker.DealerBalanceUpdater:27"
	var tail string = "Starting update of dealer balance"
	var line string = expectedSource + " - " + tail

	leftOver, src := removeSource(line)
	if src != expectedSource {
		t.Errorf("Expected source to be extracted as [%v] but got [%v]", expectedSource, src)
	}
	if leftOver != tail {
		t.Errorf("Expected after successful source removal that line be [%v] instead got [%v]", tail, leftOver)
	}

	line = "String with no dash"
	leftOver, src = removeSource(line)
	if src != "" {
		t.Errorf("When line contains no dash and therefore there is no source, source should be empty. Got [%v]", src)
	}
	if leftOver != line {
		t.Errorf("When no source is in given line, orignal line should be returned. Got [%v]", leftOver)
	}
}

func TestExtractLogLevel(t *testing.T) {
	var infoLine string = "  INFO someother stuff here"
	var debugLine string = " DEBUG someother stuff here"
	var traceLine string = " TRACE someother stuff here"
	var errorLine string = "ERROR someother stuff here"
	var leftOver = "someother stuff here"

	line, level := removeLogLevel(infoLine)
	if line != leftOver {
		t.Errorf("Expected line with INFO removed to be returned got [%v]", line)
	}
	if level != INFO_LOG_LEVEL {
		t.Errorf("Expected INFO LogLevel to be removeed got [%v]", level)
	}

	line, level = removeLogLevel(debugLine)
	if line != leftOver {
		t.Errorf("Expected line with DEBUG removed to be returned got [%v]", line)
	}
	if level != DEBUG_LOG_LEVEL {
		t.Errorf("Expected DEBUG LogLevel to be removeed got [%v]", level)
	}

	line, level = removeLogLevel(traceLine)
	if line != leftOver {
		t.Errorf("Expected line with TRACE removed to be returned got [%v]", line)
	}
	if level != TRACE_LOG_LEVEL {
		t.Errorf("Expected TRACE LogLevel to be removeed got [%v]", level)
	}

	line, level = removeLogLevel(errorLine)
	if line != leftOver {
		t.Errorf("Expected line with ERROR removed to be returned got [%v]", line)
	}
	if level != ERROR_LOG_LEVEL {
		t.Errorf("Expected ERROR LogLevel to be removeed got [%v]", level)
	}

	var unkownLine = "UNKOWN something"
	line, level = removeLogLevel(unkownLine)
	if level != EMPTY_LOG_LEVEL {
		t.Errorf("When Log Level is not known in string or not found, EMPTY_LOG_LEVEL should be returned. Got [%v]", level)
	}
	if line != unkownLine {
		t.Errorf("When Log Level is unknown, original line should be returned got [%v]", line)
	}

}

func TestToTimestamp(t *testing.T) {
	var date string = "2016-03-23 15:15:15,155"
	var expectedTime time.Time = time.Date(2016, 03, 23, 15, 15, 15, 155*1000000, time.UTC)

	timestamp, err := toTimestamp(date)
	if err != nil {
		t.Errorf("Expected [%v] to be parsed without an error. Got [%v]", err)
	}
	if *timestamp != expectedTime {
		t.Errorf("Expected string: [%v] to be parsed as [%v]", date, expectedTime)
	}

	timestamp, err = toTimestamp("invalid date")
	if err == nil {
		t.Errorf("Expected Error to not be nil when invalid date is given to be converted")
	}
}

func TestRemoveDatePortion(t *testing.T) {
	var partBetweenBrackets string = "2016-01-01 15:15:15,155"
	var part = "[" + partBetweenBrackets + "]"

	newPart, removedPart := removeFirstBetweenBrackets(part)
	if removedPart != partBetweenBrackets {
		t.Errorf("Expected [%v] to be returned as the part between brackets but got [%v]", partBetweenBrackets, removedPart)
	}
	if newPart != "" {
		t.Errorf("Expected returned string to not contain the removed part anymore, but got [%v]", newPart)
	}

	part = "No brakcets"
	newPart, removedPart = removeFirstBetweenBrackets(part)
	if newPart != part {
		t.Errorf("Expected original string to be returned when given string does not contain brackets, got [%v]", newPart)
	}
	if removedPart != "" {
		t.Error("When given string contains no brackets, removed Part should be empty. Got [%v]", removedPart)
	}
}