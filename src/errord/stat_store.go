package errord

import (
	"database/sql"
	"log"
	"time"
)

type StatStore interface {
	GetStatItem(excp string) *StatItem
	InsertOrUpdateStatItem(s *StatItem) error
	FetchSummaries() []Summary
	FetchDaySummaries() []DaySummary
	GetDaySummary(e *ErrorEvent) *DaySummary
	UpdateDaySummaries() error
}

type statStore struct {
	db *sql.DB
}

func (store *statStore) GetStatItem(excp string) *StatItem {
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

func (store *statStore) InsertOrUpdateStatItem(s *StatItem) error {
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

func (store *statStore) FetchSummaries() []Summary {
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

func (store *statStore) FetchDaySummariesByException(excp string) []*DaySummary {
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

func (store *statStore) FetchDaySummaries() []DaySummary {
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

func (store *statStore) GetDaySummary(event *ErrorEvent) *DaySummary {
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

func (store *statStore) UpdateDaySummaries() error {
	_, err := store.db.Exec(`
		insert or ignore into error_day_summary(error_date, exception, total) select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events group by DATE(error_date), exception
	`)
	return err
}

func toDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, date)
}

func toDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
