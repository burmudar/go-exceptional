package errord

import (
	"database/sql"
	"log"
	"time"
)

type StatStore interface {
	GetStatItem(name string) *StatItem
	InsertOrUpdateStatItem(s *StatItem) error
	FetchSummaries() []Summary
	FetchDaySummaries() []DaySummary
	GetDaySummary(e *ErrorEvent) *DaySummary
	UpdateDaySummaries() error
}

type statStore struct {
	db *sql.DB
}

func (store *statStore) GetStatItem(name string) *StatItem {
	r := store.db.QueryRow(`select name, mean, variance, std_dev, total, day_count, modified_at from event_stats where name = ?`, name)
	i := new(StatItem)
	var tempDate string
	err := r.Scan(&i.Name, &i.Mean, &i.Variance, &i.StdDev, &i.Total, &i.DayCount, &tempDate)
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
	_, err := store.db.Exec(`insert into event_stats(name, mean, variance, std_dev, total, day_count, modified_at) values (?, ?, ?, ?, ?, ?, ?)`,
		&s.Name, &s.Mean, &s.Variance, &s.StdDev, &s.Total, &s.DayCount, &date)
	if err != nil {
		log.Println("Assuming insert failed because record already exists trying UPDATE")
		_, err := store.db.Exec(`UPDATE event_stats SET mean = ?, variance = ?, std_dev = ?, total = ?, day_count = ?, modified_at = ? WHERE name = ?`,
			&s.Mean, &s.Variance, &s.StdDev, &s.Total, &s.DayCount, &date, &s.Name)
		if err != nil {
			log.Println("Failed UPDATING DB with [%v] : %v\n", *s, err)
		}
	}
	return err
}

func (store *statStore) FetchSummaries() []Summary {
	var summaries []Summary
	rows, err := store.db.Query("select min(created_at) as first_seen, name, sum(total) as total from day_summary group by name order by created_at")
	if err != nil {
		log.Printf("Failed to retrieve all Summaries: %v\n", err)
		return summaries
	}
	for rows.Next() {
		var s Summary
		var tempDate string
		err = rows.Scan(&tempDate, &s.Name, &s.Total)
		if err != nil {
			log.Printf("Failed mapping summary: %v", err)
		} else {
			s.StartDate, err = toDate(tempDate)
			if err != nil {
				log.Printf("Unkown Date format in Day Summary: %v", err)
			} else {
				s.EndDate = time.Now()
				s.DaySummaries = store.FetchDaySummariesByName(s.Name)
				summaries = append(summaries, s)
			}
		}
	}
	return summaries
}

func (store *statStore) FetchDaySummariesByName(name string) []*DaySummary {
	var summaries []*DaySummary
	rows, err := store.db.Query("select * from day_summary where name = ?", name)
	if err != nil {
		log.Printf("Failed fetching Day Summaries for [%v]: %v", name, err)
		return summaries
	}
	for rows.Next() {
		var s DaySummary
		err = rows.Scan(&s.Id, &s.Date, &s.Name, &s.Total)
		if err != nil {
			log.Printf("Failed mapping Day Summary for [%v]: %v", name, err)
		} else {
			summaries = append(summaries, &s)
		}
	}
	log.Printf("Found and mapped %v Day Summaries for [%v]", len(summaries), name)
	return summaries
}

func (store *statStore) FetchDaySummaries() []DaySummary {
	var summaries []DaySummary
	rows, err := store.db.Query("select * from day_summary")
	if err != nil {
		return summaries
	}
	for rows.Next() {
		var s DaySummary
		rows.Scan(&s.Id, &s.Date, &s.Name, &s.Total)
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
		event.Timestamp).Scan(&tempDate, &s.Name, &s.Total)
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
		insert or ignore into day_summary(created_at, name, total) select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events group by DATE(error_date), exception
	`)
	return err
}

func toDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, date)
}

func toDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
