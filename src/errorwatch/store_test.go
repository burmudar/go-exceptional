package errorwatch

import (
	"testing"
)

func TestErrorsReturnsErrorStore(t *testing.T) {
	store := NewStore().Errors()

	if store == nil {
		t.Errorf("Invalid ErrorStore returned: %v\n", store)
	}
}

func TestStatsReturnsStatsStore(t *testing.T) {
	store := NewStore().Stats()

	if store == nil {
		t.Errorf("Invalid StatsStore returned: %v\n", store)
	}
}

func TestNotificationsReturnsStatsStore(t *testing.T) {
	store := NewStore().Notifications()

	if store == nil {
		t.Errorf("Invalid NotificationsStore returned: %v\n", store)
	}
}
