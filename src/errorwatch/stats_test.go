package errorwatch

import (
	"testing"
	"time"
)

func TestStatCacheShouldResetIsTrueWhenEventIsPastStart(t *testing.T) {
	cache := statCache{start: newTime(2016, 3, 31, 12, 0, 0)}
	event := &ErrorEvent{}
	event.Timestamp = newTime(2016, 4, 1, 12, 0, 0)
	if cache.shouldReset(event) == false {
		t.Errorf("ShouldReset should be true when ErrorEvent is past Cache start date")
	}

	event.Timestamp = newTime(2017, 3, 31, 12, 0, 0)
	if cache.shouldReset(event) == false {
		t.Errorf("ShouldReset should be true when ErrorEvent is far in the future ie. Next year")
	}

	event.Timestamp = newTime(2016, 3, 30, 12, 0, 0)
	if cache.shouldReset(event) {
		t.Errorf("ShouldReset should be false when ErrorEvent is in the past")
	}

	event.Timestamp = newTime(2016, 3, 31, 13, 0, 0)
	if cache.shouldReset(event) {
		t.Errorf("ShouldReset should be false when ErrorEvent is on the same day as cache start")
	}
}

func newTime(y, m, d, h, mm, s int) *time.Time {
	temp := time.Date(y, time.Month(m), d, h, mm, s, 0, time.Local)
	return &temp
}
