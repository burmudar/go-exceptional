package errorwatch

import (
	"testing"
	"time"
)

func TestStatCacheShouldResetIsTrueWhenEventIsPastStart(t *testing.T) {
	cache := statCache{start: newTime(2016, 3, 31, 12, 0, 0)}
	resetTime := newTime(2016, 4, 1, 12, 0, 0)
	if cache.shouldReset(resetTime) == false {
		t.Errorf("ShouldReset should be true when ErrorEvent is past Cache start date")
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

func TestCreateTimeAtStartOfToday(t *testing.T) {
	start := createTimeAtStartOfToday()

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 || start.Nanosecond() != 0 {
		t.Errorf("Start Time should be at Start of day")
	}

}

func newTime(y, m, d, h, mm, s int) *time.Time {
	temp := time.Date(y, time.Month(m), d, h, mm, s, 0, time.Local)
	return &temp
}
