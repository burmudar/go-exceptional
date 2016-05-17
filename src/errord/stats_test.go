package errord

import (
	"math"
	"testing"
	"time"
)

func TestStatCacheShouldResetIsTrueWhenEventIsPastStart(t *testing.T) {
	cache := statCache{start: newTime(2016, 3, 31, 12, 0, 0)}
	resetTime := newTime(2016, 4, 1, 12, 0, 0)
	if cache.shouldReset(resetTime) == false {
		t.Errorf("ShouldReset should be true when ErrorEvent is past Cache start endDate")
	}

	resetTime = newTime(2017, 3, 31, 12, 0, 0)
	if cache.shouldReset(resetTime) == false {
		t.Errorf("ShouldReset should be true when ErrorEvent is far in the future ie. Next year")
	}

	resetTime = newTime(2016, 3, 30, 12, 0, 0)
	if cache.shouldReset(resetTime) {
		t.Errorf("ShouldReset should be false when ErrorEvent is in the past")
	}

	resetTime = newTime(2016, 3, 31, 13, 0, 0)
	if cache.shouldReset(resetTime) {
		t.Errorf("ShouldReset should be false when ErrorEvent is on the same day as cache start")
	}

	resetTime = newTime(2016, 3, 31, 23, 59, 59)
	if cache.shouldReset(resetTime) {
		t.Errorf("ShouldReset should be false when ErrorEvent is on the same day and close to midnight as cache start")
	}
}

func TestCalcDaysInPeriod(t *testing.T) {
	startDate := newTime(2016, 03, 30, 12, 00, 00)
	endDate := newTime(2016, 3, 31, 1, 0, 0)
	summary := Summary{
		Name:         "excp1",
		StartDate:    *startDate,
		EndDate:      *endDate,
		Total:        10,
		DaySummaries: []*DaySummary{},
	}
	days := summary.DaysInPeriod()
	expectedDays := 1
	if expectedDays != days {
		t.Errorf("Incorrect number of days returned. Regardless of time, if its the next day, days passed should be 1: Got %v Expected %v", days, expectedDays)
	}

	endDate = startDate
	summary.EndDate = *endDate
	days = summary.DaysInPeriod()
	expectedDays = 0
	if expectedDays != days {
		t.Errorf("Incorrect number of days returned. If endDate given is same as endDate first seen days should be 0: Got %v Expected %v", days, expectedDays)
	}

	endDate = newTime(2016, 4, 30, 12, 12, 12)
	summary.EndDate = *endDate
	expectedDays = 31
	days = summary.DaysInPeriod()
	if expectedDays != days {
		t.Errorf("Incorrect number of days returned. Got %v Expected %v", days, expectedDays)
	}
}

func TestCreateTimeAtStartOfToday(t *testing.T) {
	start := createTimeAtStartOfToday()

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 || start.Nanosecond() != 0 {
		t.Errorf("Start Time should be at Start of day")
	}

}

func TestCalcStatsWithNoSummaries(t *testing.T) {
	engine := statEngine{}
	summaries := []Summary{}

	stats := engine.calcStats(summaries)
	if len(stats) != 0 {
		t.Errorf("When there are no summaries, returned stats array should be empty")
	}
}

func TestCalcStatsWithSummaries(t *testing.T) {
	engine := statEngine{}
	day1 := newTime(2016, 03, 31, 12, 00, 00)

	day2 := newTime(2016, 04, 01, 12, 00, 00)
	summary1 := Summary{
		Name:      "excp1",
		StartDate: *day1,
		EndDate:   *day2,
		Total:     10,
		DaySummaries: []*DaySummary{
			&DaySummary{1, *day1, "excp1", 5},
			&DaySummary{2, *day2, "excp1", 5},
		},
	}
	summary2 :=
		Summary{
			Name:      "excp2",
			StartDate: *day1,
			EndDate:   *day2,
			Total:     12,
			DaySummaries: []*DaySummary{
				&DaySummary{3, *day1, "excp2", 6},
				&DaySummary{4, *day2, "excp2", 6},
			},
		}
	summaries := []Summary{summary1, summary2}
	stats := engine.calcStats(summaries)

	if len(stats) != 2 {
		t.Errorf("Should return 2 stat items, since out of 4 summaires, 2 exceptions are unique")
	}

	for i, stat := range stats {
		if stats[i].Name != "excp1" && stats[i].Name != "excp2" {
			t.Errorf("Stat Item at [%v] contains incorrect Name: %v", i, stats[i].Name)
		}
		if stats[i].Name == "excp1" && stats[i].Total != 10 {
			t.Errorf("Excp1 has 5 errors for day 1 and 5 for day 2, therefore Total Errors should be 10")
		}
		if stats[i].Name == "excp2" && stats[i].Total != 12 {
			t.Errorf("Excp2 has 6 errors for day 1 and 6 for day 2, therefore Total Errors should be 12")
		}
		if stat.DayCount != 1 {
			t.Errorf("Each excp happens on two separate days, therefore DayCount of statItem should be 2. Got %v", stat.DayCount)
		}
	}

}

func TestCalcAvg(t *testing.T) {
	day1 := newTime(2016, 03, 31, 12, 00, 00)
	var day2 = time.Now()
	summaries := []*DaySummary{
		&DaySummary{1, *day1, "excp1", 5},
		&DaySummary{2, day2, "excp1", 5},
	}
	summary := Summary{}
	summary.Name = "excp1"
	summary.StartDate = *day1
	summary.EndDate = day2
	summary.Total = 10
	summary.DaySummaries = summaries
	knownAvg := float64(summary.Total / summary.DaysInPeriod())

	avg := summary.calcAvg()
	if avg != knownAvg {
		t.Errorf("Incorrect Avg calculated for summaries")
	}

	summary = Summary{}
	avg = summary.calcAvg()
	if avg != 0 {
		t.Errorf("Avg should be 0 for empty summaries")
	}
}

func TestCalcVariance(t *testing.T) {
	day1 := newTime(2016, 03, 31, 12, 00, 00)
	var day2 = time.Now()
	summaries := []*DaySummary{
		&DaySummary{1, *day1, "excp1", 5},
		&DaySummary{2, day2, "excp1", 5},
	}
	summary := Summary{}
	summary.Name = "excp1"
	summary.StartDate = *day1
	summary.EndDate = day2
	summary.Total = 10
	summary.DaySummaries = summaries
	knownAvg := float64(summary.Total / summary.DaysInPeriod())
	knownVariance := (math.Pow(float64(summary.DaySummaries[0].Total)-knownAvg, 2) + math.Pow(float64(summary.DaySummaries[0].Total)-knownAvg, 2)) / float64(summary.DaysInPeriod())

	variance := summary.calcVariance(knownAvg)
	if variance != knownVariance {
		t.Errorf("Incorrect Variance calculated for summaries. Got %v Expected %v", variance, knownVariance)
	}

	summary = Summary{}
	variance = summary.calcVariance(knownAvg)
	if variance != 0 {
		t.Errorf("Variance should be 0 for empty summaries")
	}
}

func newTime(y, m, d, h, mm, s int) *time.Time {
	temp := time.Date(y, time.Month(m), d, h, mm, s, 0, time.Local)
	return &temp
}
