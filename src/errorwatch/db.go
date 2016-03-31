package errorwatch

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
)

const SQL_TABLE_ERROR_EVENTS string = `create table error_events
	(
		id INTEGER not null primary key,
		event_datetime DATETIME not null,
		level VARCHAR(10) not null,
		source VARCHAR(30) not null,
		description VARCHAR(255) not null,
		exception VARCHAR(255) not null,
		excp_description VARCHAR(255) not null,
		unique(event_datetime, source, description)
	)
	`
const SQL_TABLE_NOTIFICATIONS string = `create table notifications(
		id INTEGER not null primary key,
		created_at DATETIME not null,
		exception VARCHAR(255) not null,
		unique(created_at, exception))`

const SQL_ERROR_STATS string = `
	create table error_stats (
		id INTEGER not null primary key, 
		exception VARCHAR(255),
		mean DOUBLE not null,
		variance INTEGER not null,
		std_dev DOUBLE not null,
		total_errs INTEGER not null,
		day_count INTEGER not null,
		modified_at DATETIME not null,
		unique(exception)
	)
	`
const SQL_TABLE_ERROR_DAY_SUMMARY string = `
	create table error_day_summary(
		id INTEGER not null primary key,
		error_date DATETIME not null,
		exception VARCHAR(255) not null,
		total INTEGER not null,
		unique(error_date, exception)
	)
	`

var ErrTableExists error = errors.New("Not creating Table. Table already exists")

type ErrorWatchStore interface {
	InsertOrUpdateStatItem(s *StatItem) error
	FetchDaySummaries() []Summary
	GetDaySummary(e *ErrorEvent) *Summary
	UpdateDaySummaries() error
	AddErrorEvent(e *ErrorEvent) error
}

type dbStore struct {
	db *sql.DB
}

func NewErrorWatchStore() ErrorWatchStore {
	db, errors := initDB()
	if len(errors) != 0 {
		log.Panicf("Failed initializing database: %v\n", errors)
	}
	dbStore := &dbStore{db}
	return dbStore
}

func (store *dbStore) InsertOrUpdateStatItem(s *StatItem) error {
	var date = s.ModifiedAt.Format(DATE_FORMAT)
	_, err := store.db.Exec(`insert into error_stats(exception, mean, variance, std_dev, total_errs, day_count, modified_at) values (?, ?, ?, ?, ?, ?, ?)`,
		&s.Exception, &s.Mean, &s.Variance, &s.StdDev, &s.TotalErrors, &s.DayCount, &date)
	if err != nil {
		log.Println("Assuming insert failed because record already exists trying UPDATE")
		_, err := store.db.Exec(`UPDATE error_stats SET mean = ?, variance = ?, std_dev = ?, total_errs = ?, day_count = ?, modified_at = ? WHERE exception = ?`,
			&s.Mean, &s.Variance, &s.StdDev, &s.TotalErrors, &s.DayCount, &date, &s.Exception)
		if err != nil {
			log.Println("Failed UPDATING DB with [%v] : %v\n", *s, err)
		}
	}
	return err
}

func (store *dbStore) FetchDaySummaries() []Summary {
	var summaries []Summary
	rows, err := store.db.Query("select * from error_day_summary")
	if err != nil {
		return summaries
	}
	for rows.Next() {
		var s Summary
		rows.Scan(&s.Id, &s.Date, &s.Exception, &s.Total)
		summaries = append(summaries, s)
	}
	return summaries
}

func (store *dbStore) GetDaySummary(event *ErrorEvent) *Summary {
	s := new(Summary)
	var tempDate string
	/*
		Scan into tempDate string since Scan can't automatically figure out the Date format. So we scan to a string and parse the string with a known date layout
	*/
	err := store.db.QueryRow("select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events where error_date = DATE(?) group by DATE(error_date), exception",
		event.Timestamp).Scan(&tempDate, &s.Exception, &s.Total)
	if err != nil {
		log.Printf("Failed to map Day Summary for [%v] : %v\n", *event, err)
	}
	date, err := toDate(tempDate)
	if err != nil {
		log.Printf("Unknown Date format: %v", tempDate)
	}
	s.Date = date
	return s
}

func (store *dbStore) UpdateDaySummaries() error {
	_, err := store.db.Exec(`
		insert or ignore into error_day_summary(error_date, exception, total) select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events group by DATE(error_date), exception
	`)
	return err
}

func (store *dbStore) AddErrorEvent(e *ErrorEvent) error {
	var count int
	store.db.QueryRow(`select count(id) from error_events where event_datetime=? AND source=? AND description=? AND exception=? AND excp_description=?`,
		e.Timestamp, e.Source, e.Description, e.Exception, e.Description).Scan(&count)
	if count > 0 {
		log.Printf("[%v : %v] Already exists!\n", *e.Timestamp, e.Exception)
		return nil
	}
	_, err := store.db.Exec(`insert into error_events(event_datetime, level, source, description, exception, excp_description) 
	values (?, ?, ?, ?, ?, ?)`, e.Timestamp, string(e.Level), e.Source, e.Description, e.Exception, e.Description)
	if err != nil {
		return err
	}
	return nil
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
	tables["error_stats"] = SQL_ERROR_STATS
	tables["error_day_summary"] = SQL_TABLE_ERROR_DAY_SUMMARY
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
	err := db.QueryRow("select name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&table)
	table = strings.Trim(table, " ")
	if err == sql.ErrNoRows || table == "" {
		return false
	} else {
		return true
	}

}

func toDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, date)
}

func toDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
