package errord

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

const SQL_TABLE_ERROR_EVENTS string = `create table error_events
	(
		id INTEGER not null primary key,
		event_datetime DATETIME not null,
		level VARCHAR(10) not null,
		description VARCHAR(255) not null,
		exception VARCHAR(255) not null,
		excp_description VARCHAR(255) not null,
		unique(event_datetime, exception)
	)
	`
const SQL_TABLE_NOTIFICATIONS string = `create table notifications(
		id INTEGER not null primary key,
		created_at DATETIME not null,
		subject VARCHAR(255) not null,
		unique(created_at, subject))`

const SQL_EVENT_STATS string = `
	create table event_stats (
		id INTEGER not null primary key, 
		name VARCHAR(255),
		mean DOUBLE not null,
		variance INTEGER not null,
		std_dev DOUBLE not null,
		total INTEGER not null,
		day_count INTEGER not null,
		modified_at DATETIME not null,
		unique(name)
	)
	`
const SQL_TABLE_DAY_SUMMARY string = `
	create table day_summary(
		id INTEGER not null primary key,
		created_at DATETIME not null,
		name VARCHAR(255) not null,
		count INTEGER not null,
		total INTEGER not null,
		unique(created_at, name)
	)
	`

var ErrTableExists error = errors.New("Not creating Table. Table already exists")

type Store interface {
	Init() []error
	Errors() ErrorStore
	Metrics() MetricStore
	Stats() StatStore
	Notifications() NotifyStore
}

type dbStore struct {
	db *sql.DB
}

func NewStore() Store {
	s := new(dbStore)
	return s
}

func (s *dbStore) Init() []error {
	db, errors := initDB()
	s.db = db
	return errors
}

func (s *dbStore) Errors() ErrorStore {
	return &errorStore{s.db}
}

func (s *dbStore) Metrics() MetricStore {
	return &metricStore{s.db}
}

func (s *dbStore) Notifications() NotifyStore {
	return &notifyStore{s.db}
}
func (s *dbStore) Stats() StatStore {
	return &statStore{s.db}
}

func createTable(db *sql.DB, table string, sql string) error {
	var err error
	if hasTable(db, table) {
		return ErrTableExists
	} else {
		_, err = db.Exec(sql)
		return err
	}
}

func initDB() (*sql.DB, []error) {
	var errors []error = []error{}
	db, err := sql.Open("sqlite3", "errors.db")

	if err != nil {
		return nil, []error{err}
	}
	var tables map[string]string = make(map[string]string)
	tables["error_events"] = SQL_TABLE_ERROR_EVENTS
	tables["event_stats"] = SQL_EVENT_STATS
	tables["day_summary"] = SQL_TABLE_DAY_SUMMARY
	tables["notifications"] = SQL_TABLE_NOTIFICATIONS
	for table, sql := range tables {
		err := createTable(db, table, sql)
		if err == ErrTableExists {
			log.Printf("Database Initialize Error: [%v]\n", err)
		} else if err != nil {
			errors = append(errors, err)
		}
	}
	return db, errors
}

func hasTable(db *sql.DB, name string) bool {
	var table string
	err := db.QueryRow("select name FROM sqlite_master WHERE (type='table' OR type='view') AND name=?", name).Scan(&table)
	table = strings.Trim(table, " ")
	if err == sql.ErrNoRows || table == "" {
		return false
	} else {
		return true
	}

}
