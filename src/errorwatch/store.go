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
const SQL_VIEW_EXCEPTIONS_PER_DAY string = `CREATE VIEW exceptions_per_day AS select DATE(event_datetime), exception, count(exception) as excp_count from error_events group by DATE(event_datetime),exception order by event_datetime`
const SQL_VIEW_UNIQUE_EXCEPTIONS string = `CREATE VIEW unique_exception AS select min(date(event_datetime)), exception, count(exception) as excp_count, cast((julianday('now') - julianday(event_datetime)) as int) as total from error_events group by exception order by event_datetime`

var ErrTableExists error = errors.New("Not creating Table. Table already exists")

type Store interface {
	Init() []error
	Errors() ErrorWatchStore
	Stats() StatStore
	Notifications() NotifyStore
}

type ErrorWatchStore interface {
	AddErrorEvent(e *ErrorEvent) error
}

type StatStore interface {
	GetStatItem(excp string) *StatItem
	InsertOrUpdateStatItem(s *StatItem) error
	FetchSummaries() []Summary
	FetchDaySummaries() []DaySummary
	GetDaySummary(e *ErrorEvent) *DaySummary
	UpdateDaySummaries() error
}

type NotifyStore interface {
	UpdateNotificationSent(e *ErrorEvent) error
	HasNotification(e *ErrorEvent) bool
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

func (s *dbStore) Errors() ErrorWatchStore {
	return s
}

func (s *dbStore) Notifications() NotifyStore {
	return s
}
func (s *dbStore) Stats() StatStore {
	return s
}

func (s *dbStore) UpdateNotificationSent(e *ErrorEvent) error {
	_, err := s.db.Exec("insert into notifications(created_at, exception) values(DATE(?), ?)", time.Now(), e.Exception)
	return err
}

func (s *dbStore) HasNotification(e *ErrorEvent) bool {
	r := s.db.QueryRow(`select count(*) from notifications where created_at = DATE(?) and exception = ?`, time.Now(), e.Exception)
	var count int
	err := r.Scan(&count)
	if err != nil {
		log.Printf("Failed mapping notification count: %v\n", err)
		return true
	}
	return count > 0
}

func (store *dbStore) GetStatItem(excp string) *StatItem {
	r := store.db.QueryRow(`select exception, mean, variance, std_dev, total_errs, day_count, modified_at from error_stats where exception = ?`, excp)
	i := new(StatItem)
	var tempDate string
	err := r.Scan(&i.Exception, &i.Mean, &i.Variance, &i.StdDev, &i.TotalErrors, &i.DayCount, &tempDate)
	if err != nil {
		log.Printf("Failed mapping stat item: %v\n", err)
	}
	date, err := toDateTime(tempDate)
	i.ModifiedAt = &date
	if err != nil {
		log.Printf("Failed parsing date: %v : %v\n", tempDate, err)
	}
	return i
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

func (store *dbStore) FetchSummaries() []Summary {
	var summaries []Summary
	rows, err := store.db.Query("select min(error_date) as first_seen, exception, sum(total) as total from error_day_summary group by exception order by error_date")
	if err != nil {
		log.Printf("Failed to retrieve all Summaries: %v\n", err)
		return summaries
	}
	for rows.Next() {
		var s Summary
		var tempDate string
		err = rows.Scan(&tempDate, &s.Exception, &s.Total)
		if err != nil {
			log.Printf("Failed mapping summary: %v", err)
		} else {
			s.StartDate, err = toDate(tempDate)
			if err != nil {
				log.Printf("Unkown Date format in Day Summary: %v", err)
			} else {
				s.EndDate = time.Now()
				s.DaySummaries = store.FetchDaySummariesByException(s.Exception)
				summaries = append(summaries, s)
			}
		}
	}
	return summaries
}

func (store *dbStore) FetchDaySummariesByException(excp string) []*DaySummary {
	var summaries []*DaySummary
	rows, err := store.db.Query("select * from error_day_summary where exception = ?", excp)
	if err != nil {
		log.Printf("Failed fetching Day Summaries for [%v]: %v", excp, err)
		return summaries
	}
	for rows.Next() {
		var s DaySummary
		err = rows.Scan(&s.Id, &s.Date, &s.Exception, &s.Total)
		if err != nil {
			log.Printf("Failed mapping Day Summary for [%v]: %v", excp, err)
		} else {
			summaries = append(summaries, &s)
		}
	}
	log.Printf("Found and mapped %v Day Summaries for [%v]", len(summaries), excp)
	return summaries
}

func (store *dbStore) FetchDaySummaries() []DaySummary {
	var summaries []DaySummary
	rows, err := store.db.Query("select * from error_day_summary")
	if err != nil {
		return summaries
	}
	for rows.Next() {
		var s DaySummary
		rows.Scan(&s.Id, &s.Date, &s.Exception, &s.Total)
		summaries = append(summaries, s)
	}
	return summaries
}

func (store *dbStore) GetDaySummary(event *ErrorEvent) *DaySummary {
	s := new(DaySummary)
	var tempDate string
	/*
		Scan into tempDate string since Scan can't automatically figure out the Date format. So we scan to a string and parse the string with a known date layout
	*/
	err := store.db.QueryRow("select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events where error_date = DATE(?) group by DATE(error_date), exception",
		event.Timestamp).Scan(&tempDate, &s.Exception, &s.Total)
	if err != nil {
		log.Printf("Failed to map DaySummary for [%v] : %v\n", *event, err)
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
	log.Printf("Inserting -> %v : %v\n", *e.Timestamp, e.Exception)
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
	//kind of like tables ... but really views
	tables["exceptions_per_day"] = SQL_VIEW_EXCEPTIONS_PER_DAY
	tables["unique_exceptions"] = SQL_VIEW_UNIQUE_EXCEPTIONS
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

func toDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, date)
}

func toDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
