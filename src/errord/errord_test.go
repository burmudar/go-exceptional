package errord

import (
	"testing"
	"time"
)

func TestparseLogLineWithEmptyLine(t *testing.T) {
	_, err := parseLogLine("")

	if err == nil {
		t.Errorf("When line is empty. Error should be returned")
	}
}

func TestparseLogLineOfINFOLine(t *testing.T) {
	const INFO_LINE string = "[2016-03-23 15:41:48,564] INFO  worker.DealerBalanceUpdater:27 - Starting update of dealer balance"
	logEvent, err := parseLogLine(INFO_LINE)

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

	expectedDescription := "Starting update of dealer balance"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the INFO line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestparseLogLineOfTRACELine(t *testing.T) {
	const TRACE_LINE string = "[2016-03-23 15:41:48,608] TRACE worker.SimConsumerImpl:183 - null : Retrieving Balance of Sim id: 0 from Recharge Service"
	logEvent, err := parseLogLine(TRACE_LINE)

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

	expectedDescription := "null : Retrieving Balance of Sim id: 0 from Recharge Service"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the TRACE line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseLogLineOfDEBUGLine(t *testing.T) {
	const DEBUG_LINE string = "[2016-03-23 15:41:48,615] DEBUG worker.SimConsumerImpl:129 - null : Sending recharge to service"
	logEvent, err := parseLogLine(DEBUG_LINE)

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

	expectedDescription := "null : Sending recharge to service"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the DEBUG line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
	}
}

func TestParseLogLineOfERRORLine(t *testing.T) {
	const ERROR_LINE string = "[2016-03-23 15:41:48,939] ERROR client.AirtelService:54 - 0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	logEvent, err := parseLogLine(ERROR_LINE)

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

	expectedDescription := "0833574730 : Encountered an error while querying balance : TranRef[testRef]"
	if logEvent.Description != expectedDescription {
		t.Errorf("Event does not contain description as its defined in the ERROR line. Got [%v] Expected [%v]", logEvent.Description, expectedDescription)
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

func TestContainsCausedBy(t *testing.T) {
	if containsCausedBy("caused by:") {
		t.Errorf("Should only return true when string contains 'Caused by:' case sensitive")
	}
	if containsCausedBy("") {
		t.Errorf("Empty string does not contain 'Caused by:'")
	}
	if !containsCausedBy("      Caused by:") {
		t.Errorf("String with whitespace before 'Caused by:' is valid and should not be rejected")
	}
	if !containsCausedBy("Caused by:") {
		t.Errorf("String containing exact 'Caused by:' should not be rejected")
	}

}

func TestHasCausedBy(t *testing.T) {
	event := new(ErrorEvent)

	if event.hasCausedBy() {
		t.Errorf("Should be false when exception and detail are empty")
	}

	event.Exception = "some exception"
	if !event.hasCausedBy() {
		t.Errorf("Should be true when exception has a value")
	}

	event.Detail = "some detail"
	if !event.hasCausedBy() {
		t.Errorf("Should be true when exception and detail have values")
	}
}

func TestParsedCausedBy(t *testing.T) {
	var NORMAL_CAUSED_BY string = "Caused by: com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException: UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
	var CAUSED_BY_WITH_MULTIPLE_COLONS string = "Caused by: com.flickswitch.sc.provider.ProviderServiceException: com.flickswitch.client.airtelke.balance.BalanceServiceException: Server responded with Failure status. Error detail -> TWSS_109 : Your request for service could not be processed due to Network error. Pls try again"
	var CAUSED_BY_WITHOUT_DETAIL string = "Caused by: javax.xml.bind.UnmarshalException"
	var CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL string = "Caused by:"

	excp, detail, err := parseCausedBy(NORMAL_CAUSED_BY)
	if excp == "" && detail == "" {
		t.Errorf("No Caused By extracted from valid Caused by line: [%v]. Got %v | %v", NORMAL_CAUSED_BY, excp, detail)
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITHOUT_DETAIL)

	expectedException := "javax.xml.bind.UnmarshalException"
	if excp != expectedException {
		t.Errorf("Incorrect exception extracted. Expected: [%v] got [%v]\n", expectedException, excp)
	}
	expectedDetail := ""
	if detail != expectedDetail {
		t.Errorf("Incorrect detail extracted. Expected: [%v] got [%v]\n", expectedDetail, detail)
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL)
	if err == nil {
		t.Errorf("Caused by with no exception nor detail is invalid and should return err. Got [%v]\n", err)
	}
	if excp != "" && detail != "" {
		t.Errorf("Exception and detail should be empty strings on caused by with no exception or detail")
	}

	excp, detail, err = parseCausedBy("")
	if err == nil {
		t.Errorf("Empty string is not a valid Caused By Error should not be nil")
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITH_MULTIPLE_COLONS)
	if excp != "com.flickswitch.sc.provider.ProviderServiceException" {
		t.Errorf("Incorrect exception extracted from Caused by with multiple colons")
	}
	expectedDetail = "com.flickswitch.client.airtelke.balance.BalanceServiceException: Server responded with Failure status. Error detail -> TWSS_109 : Your request for service could not be processed due to Network error. Pls try again"
	if detail != expectedDetail {
		t.Errorf("Incorrect detail extracted from Caused by with multiple colons. [%v] != [%v]\n", expectedDetail, detail)
	}
}

func TestCreateErrorEvent(t *testing.T) {
	var NORMAL_CAUSED_BY string = "Caused by: com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException: UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
	var CAUSED_BY_WITH_MULTIPLE_COLONS string = "Caused by: com.flickswitch.sc.provider.ProviderServiceException: com.flickswitch.client.airtelke.balance.BalanceServiceException: Server responded with Failure status. Error detail -> TWSS_109 : Your request for service could not be processed due to Network error. Pls try again"
	var CAUSED_BY_WITHOUT_DETAIL string = "Caused by: javax.xml.bind.UnmarshalException"
	var CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL string = "Caused by:"
	var event *Event = new(Event)

	excp, detail, err := parseCausedBy(NORMAL_CAUSED_BY)
	errorEvent, err := createErrorEvent(NORMAL_CAUSED_BY, event)
	if excp != errorEvent.Exception && detail != errorEvent.Detail {
		t.Errorf("Exception [%v] Detail [%v] not added to event: %v", excp, detail, errorEvent)
	}
	if errorEvent.hasCausedBy() && err != nil {
		t.Errorf("Received error when adding Caused By, event though event reports it has caused by")
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITHOUT_DETAIL)
	errorEvent, err = createErrorEvent(CAUSED_BY_WITHOUT_DETAIL, event)
	if excp != errorEvent.Exception && detail != errorEvent.Detail {
		t.Errorf("Exception [%v] Detail [%v] not added to event: %v", excp, detail, errorEvent)
	}
	if errorEvent.hasCausedBy() && err != nil {
		t.Errorf("Received error when adding Caused By, event though event reports it has caused by")
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL)
	errorEvent, err = createErrorEvent(CAUSED_BY_WITHOUT_EXCEPTION_OR_DETAIL, event)
	if err == nil {
		t.Errorf("Caused by with no exception nor detail is invalid and should return err. Got [%v]\n", err)
	}
	if errorEvent != nil {
		t.Error("When Caused By is not present, no ErrorEvent should be created")
	}

	excp, detail, err = parseCausedBy("")
	errorEvent, err = createErrorEvent("", event)
	if err == nil {
		t.Errorf("Empty string is not a valid Caused By Error should not be nil")
	}
	if errorEvent != nil {
		t.Error("When Caused By is not present, no ErrorEvent should be created")
	}

	excp, detail, err = parseCausedBy(CAUSED_BY_WITH_MULTIPLE_COLONS)
	errorEvent, err = createErrorEvent(CAUSED_BY_WITH_MULTIPLE_COLONS, event)
	if excp != errorEvent.Exception && detail != errorEvent.Detail {
		t.Errorf("Exception [%v] Detail [%v] not added to event.", excp, detail, errorEvent)
	}
	if errorEvent.hasCausedBy() && err != nil {
		t.Errorf("Received error when adding Caused By, event though event reports it has caused by")
	}

}
