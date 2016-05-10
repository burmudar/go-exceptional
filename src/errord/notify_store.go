package errord

import (
	"database/sql"
	"log"
	"time"
)

type NotifyStore interface {
	UpdateNotificationSent(e *ErrorEvent) error
	HasNotification(e *ErrorEvent) bool
}

type notifyStore struct {
	db *sql.DB
}

func (s *notifyStore) UpdateNotificationSent(e *ErrorEvent) error {
	_, err := s.db.Exec("insert into notifications(created_at, exception) values(DATE(?), ?)", time.Now(), e.Exception)
	return err
}

func (s *notifyStore) HasNotification(e *ErrorEvent) bool {
	r := s.db.QueryRow(`select count(*) from notifications where created_at = DATE(?) and exception = ?`, time.Now(), e.Exception)
	var count int
	err := r.Scan(&count)
	if err != nil {
		log.Printf("Failed mapping notification count: %v\n", err)
		return true
	}
	return count > 0
}
