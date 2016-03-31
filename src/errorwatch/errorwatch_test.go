package errorwatch

import (
	"testing"
	"time"
)

func TestParseLogLineOfINFOLine(t *testing.T) {
	const INFO_LINE string = "[2016-03-23 15:41:48,564] INFO  worker.DealerBalanceUpdater:27 - Starting update of dealer balance"
	logEvent, err := ParseLogLine(INFO_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When Event is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid INFO log line should not return an error: [%v]", err)
	}

	if logEvent.Level != INFO_LOG_LEVEL {
		t.Errorf("Event should have INFO Level. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 564*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("Event does not contain the Timestamp as its defined in the INFO line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.DealerBalanceUpdater:27"
	if logEvent.Source != expctedSource {
		t.Errorf("Event does not contain source as its defined in the INFO line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "Starting update of dealer balance"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the INFO line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseLogLineOfTRACELine(t *testing.T) {
	const TRACE_LINE string = "[2016-03-23 15:41:48,608] TRACE worker.SimConsumerImpl:183 - null : Retrieving Balance of Sim id: 0 from Recharge Service"
	logEvent, err := ParseLogLine(TRACE_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When Event is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid TRACE log line should not return an error: [%v]", err)
	}

	if logEvent.Level != TRACE_LOG_LEVEL {
		t.Errorf("Event should have TRACE Level. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 608*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("Event does not contain the Timestamp as its defined in the TRACE line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.SimConsumerImpl:183"
	if logEvent.Source != expctedSource {
		t.Errorf("Event does not contain source as its defined in the TRACE line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "null : Retrieving Balance of Sim id: 0 from Recharge Service"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the TRACE line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseLogLineOfDEBUGLine(t *testing.T) {
	const DEBUG_LINE string = "[2016-03-23 15:41:48,615] DEBUG worker.SimConsumerImpl:129 - null : Sending recharge to service"
	logEvent, err := ParseLogLine(DEBUG_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When Event is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid DEBUG log line should not return an error: [%v]", err)
	}

	if logEvent.Level != DEBUG_LOG_LEVEL {
		t.Errorf("Event should have DEBUG Level. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 615*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("Event does not contain the Timestamp as its defined in the DEBUG line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "worker.SimConsumerImpl:129"
	if logEvent.Source != expctedSource {
		t.Errorf("Event does not contain source as its defined in the DEBUG line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "null : Sending recharge to service"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the DEBUG line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseLogLineOfERRORLine(t *testing.T) {
	const ERROR_LINE string = "[2016-03-23 15:41:48,939] ERROR client.AirtelService:54 - 0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	logEvent, err := ParseLogLine(ERROR_LINE)

	if logEvent == nil && err == nil {
		t.Errorf("When Event is nil, Error cannot be nil")
	}

	if err != nil {
		t.Errorf("Parsing of valid ERROR log line should not return an error: [%v]", err)
	}

	if logEvent.Level != ERROR_LOG_LEVEL {
		t.Errorf("Event should have ERROR Level. Got [%v] instead", logEvent.Level)
	}
	expectedTime := time.Date(2016, 3, 23, 15, 41, 48, 939*1000000, time.UTC)
	if *logEvent.Timestamp != expectedTime {
		t.Errorf("Event does not contain the Timestamp as its defined in the ERROR line: [%v] - Error: [%v]", logEvent.Timestamp, err)
	}

	expctedSource := "client.AirtelService:54"
	if logEvent.Source != expctedSource {
		t.Errorf("Event does not contain source as its defined in the ERROR line. Got [%v] Expected [%v]", logEvent.Source, expctedSource)
	}

	expectedDescription := "0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the ERROR line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
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

func TestExtractLevel(t *testing.T) {
	var infoLine string = "  INFO someother stuff here"
	var debugLine string = " DEBUG someother stuff here"
	var traceLine string = " TRACE someother stuff here"
	var errorLine string = "ERROR someother stuff here"
	var leftOver = "someother stuff here"

	line, level := removeLevel(infoLine)
	if line != leftOver {
		t.Errorf("Expected line with INFO removed to be returned got [%v]", line)
	}
	if level != INFO_LOG_LEVEL {
		t.Errorf("Expected INFO Level to be removeed got [%v]", level)
	}

	line, level = removeLevel(debugLine)
	if line != leftOver {
		t.Errorf("Expected line with DEBUG removed to be returned got [%v]", line)
	}
	if level != DEBUG_LOG_LEVEL {
		t.Errorf("Expected DEBUG Level to be removeed got [%v]", level)
	}

	line, level = removeLevel(traceLine)
	if line != leftOver {
		t.Errorf("Expected line with TRACE removed to be returned got [%v]", line)
	}
	if level != TRACE_LOG_LEVEL {
		t.Errorf("Expected TRACE Level to be removeed got [%v]", level)
	}

	line, level = removeLevel(errorLine)
	if line != leftOver {
		t.Errorf("Expected line with ERROR removed to be returned got [%v]", line)
	}
	if level != ERROR_LOG_LEVEL {
		t.Errorf("Expected ERROR Level to be removeed got [%v]", level)
	}

	var unkownLine = "UNKOWN something"
	line, level = removeLevel(unkownLine)
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

func TestContainsCausedBy(t *testing.T) {
	if ContainsCausedBy("caused by:") {
		t.Errorf("Should only return true when string contains 'Caused by:' case sensitive")
	}
	if ContainsCausedBy("") {
		t.Errorf("Empty string does not contain 'Caused by:'")
	}
	if !ContainsCausedBy("      Caused by:") {
		t.Errorf("String with whitespace before 'Caused by:' is valid and should not be rejected")
	}
	if !ContainsCausedBy("Caused by:") {
		t.Errorf("String containing exact 'Caused by:' should not be rejected")
	}

}

func TestExtractCausedBy(t *testing.T) {
	var NORMAL_CAUSED_BY = "Caused by: com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException: UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
	var CAUSED_BY_WITHOUT_DETAIL = "Caused by: javax.xml.bind.UnmarshalException"
	var CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL = "Caused by:"

	causedBy, err := extractCausedBy(NORMAL_CAUSED_BY)
	if causedBy == nil {
		t.Errorf("No Caused By extracted from valid Caused by line: [%v]\n", NORMAL_CAUSED_BY)
	}
	if causedBy == nil && err == nil {
		t.Errorf("When Caused By is nil, Error cannot be nil")
	}
	expectedException := "com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException"
	if causedBy.Exception != expectedException {
		t.Errorf("Incorrect exception extracted. Expected: [%v] got [%v]\n", expectedException, causedBy.Exception)
	}
	expectedDetail := "UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
	if causedBy.Detail != expectedDetail {
		t.Errorf("Incorrect detail extracted. Expected: [%v] got [%v]\n", expectedDetail, causedBy.Detail)
	}

	causedBy, err = extractCausedBy(CAUSED_BY_WITHOUT_DETAIL)

	if causedBy == nil {
		t.Errorf("No Caused By extracted from valid Caused by line: [%v]\n", CAUSED_BY_WITHOUT_DETAIL)
	}
	if causedBy == nil && err == nil {
		t.Errorf("When Caused By is nil, Error cannot be nil")
	}
	expectedException = "javax.xml.bind.UnmarshalException"
	if causedBy.Exception != expectedException {
		t.Errorf("Incorrect exception extracted. Expected: [%v] got [%v]\n", expectedException, causedBy.Exception)
	}
	expectedDetail = ""
	if causedBy.Detail != expectedDetail {
		t.Errorf("Incorrect detail extracted. Expected: [%v] got [%v]\n", expectedDetail, causedBy.Detail)
	}
	causedBy, err = extractCausedBy(CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL)
	if causedBy != nil {
		t.Errorf("Caused by with no exception nor detail is invalid and should return nil. Got [%v]\n", causedBy)
	}
	if err == nil {
		t.Errorf("Caused by with no exception nor detail should return Error containing the reason")
	}
	causedBy, err = extractCausedBy("")
	if causedBy != nil {
		t.Errorf("Empty string is not a valid Caused By. Got [%v]\n", causedBy)
	}
	if err == nil {
		t.Errorf("Empty caused by should return an error")
	}
}

func TestParseCausedBy(t *testing.T) {
	const ERROR_LINE string = "[2016-03-23 15:41:48,939] ERROR client.AirtelService:54 - 0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	var NORMAL_CAUSED_BY = "Caused by: com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException: UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
	var CAUSED_BY_WITHOUT_DETAIL = "Caused by: javax.xml.bind.UnmarshalException"
	var CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL = "Caused by:"

	event, err := ParseLogLine(ERROR_LINE)

	errorEvent, err := ParseCausedBy(NORMAL_CAUSED_BY, event)
	causedBy, _ := extractCausedBy(NORMAL_CAUSED_BY)

	if err != nil {
		t.Errorf("Failed creating ErrorEvent with Caused By from [%v]\n", NORMAL_CAUSED_BY)
	} else if errorEvent == nil {
		t.Errorf("ErrorEvent cannot be nil, if error is nil")
	}

	if errorEvent.Exception != causedBy.Exception {
		t.Errorf("ErrorEvent does not contain Caused By Exception. Expected: [%v] Got: [%v]\n", errorEvent.Exception, causedBy.Exception)
	}

	if errorEvent.Detail != causedBy.Detail {
		t.Errorf("ErrorEvent does not contain Caused By Detail. Expected: [%v] Got: [%v]\n", errorEvent.Detail, causedBy.Detail)
	}

	errorEvent, err = ParseCausedBy(CAUSED_BY_WITHOUT_DETAIL, event)
	causedBy, _ = extractCausedBy(CAUSED_BY_WITHOUT_DETAIL)

	if err != nil {
		t.Errorf("Failed creating ErrorEvent with Caused By from [%v]\n", NORMAL_CAUSED_BY)
	} else if errorEvent == nil {
		t.Errorf("ErrorEvent cannot be nil, if error is nil")
	}

	if errorEvent.Exception != causedBy.Exception {
		t.Errorf("ErrorEvent does not contain Caused By Exception. Expected: [%v] Got: [%v]\n", errorEvent.Exception, causedBy.Exception)
	}

	if errorEvent.Detail != causedBy.Detail {
		t.Errorf("ErrorEvent does not contain Caused By Detail. Expected: [%v] Got: [%v]\n", errorEvent.Detail, causedBy.Detail)
	}

	errorEvent, err = ParseCausedBy(CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL, event)
	if err == nil {
		t.Errorf("When CausedBy parsing fails, error shoudl be returned")
	} else if errorEvent != nil {
		t.Errorf("When Error is not nil, no ErrorEvent should be returned")
	}

	errorEvent, err = ParseCausedBy(NORMAL_CAUSED_BY, nil)
	if err == nil {
		t.Errorf("Error should be returned when given Event is nil")
	}

}
