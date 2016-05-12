package errord

import (
	"database/sql"
	"log"
)

type ErrorStore interface {
	Add(e *ErrorEvent) error
}

type errorStore struct {
	db *sql.DB
}

func (store *errorStore) Add(e *ErrorEvent) error {
	var count int
	log.Printf("Inserting -> %v : %v\n", *e.Timestamp, e.Exception)
	store.db.QueryRow(`select count(id) from error_events where event_datetime=? description=? AND exception=? AND excp_description=?`,
		e.Timestamp, e.Description, e.Exception, e.Detail).Scan(&count)
	if count > 0 {
		log.Printf("[%v : %v] Already exists!\n", *e.Timestamp, e.Exception)
		return nil
	}
	_, err := store.db.Exec(`insert into error_events(event_datetime, level, description, exception, excp_description) 
	values (?, ?, ?, ?, ?, ?)`, e.Timestamp, string(e.Level), e.Description, e.Exception, e.Detail)
	if err != nil {
		return err
	}
	return nil
}
