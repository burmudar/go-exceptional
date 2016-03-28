package main

import (
	"bufio"
	"database/sql"
	"errorlog"
	"errors"
	"fmt"
	"github.com/hpcloud/tail"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var db *sql.DB

var ErrNotCausedByLine error = errors.New("Line does not contain 'Caused by'")
var ErrTableExists error = errors.New("Not creating Table. Table already exists")

func main() {

	files := []string{}
	filepath.Walk("logs", func(p string, i os.FileInfo, err error) error {
		if path.Ext(p) == ".log" {
			files = append(files, p)
		}
		return nil
	})
	err := initDB()
	if err != nil {
		fmt.Printf("Failed initializing Database: [%v]\n", err)
		return
	} else {
		fmt.Println("Database initiliazed")
	}
	loadAll(files)
	initStats()
	var eventProcess chan errorlog.ErrorEvent = make(chan errorlog.ErrorEvent)
	go func() {
		var start time.Time = time.Now()
		var statCache map[string]*StatItem = make(map[string]*StatItem)
		for event := range eventProcess {
			if isEventAfterStart(&start, &event) {
				start = time.Now()
				fmt.Println("Event is day after we started. Purging Stat cache")
				statCache = make(map[string]*StatItem)
				fmt.Println("Recalculating stats")
				calcStats()
			}
			fmt.Printf("Processing: %v\n", event)
			var statItem *StatItem
			var ok bool
			if statItem, ok = statCache[event.Exception]; !ok {
				statItem = getStatItem(event.Exception)
				statCache[event.Exception] = statItem
			}
			if statItem == nil {
				notify(&event)
			} else {
				fmt.Printf("Stats: %v\n", *statItem)
				var s *Summary = getDaySummaryFor(&event)
				fmt.Printf("Summary: %v\n", *s)
				max := statItem.Mean + statItem.StdDev
				fmt.Printf("Total[%v] : Max[%v]\n", s.Total, max)
				if s.Total >= int(max) {
					notify(&event)
				} else {
				}
			}

		}
	}()
	watchFile("test.log", eventProcess)
}

func isEventAfterStart(start *time.Time, event *errorlog.ErrorEvent) bool {
	fmt.Printf("%v : %v", start.Day(), event.Timestamp.Day())
	return start.Day()-event.Timestamp.Day() > 0
}

func notify(e *errorlog.ErrorEvent) {
	fmt.Printf("NOTIFYING: [%v]\n", *e)
}

func getStatItem(excp string) *StatItem {
	r := db.QueryRow(`select exception, mean, variance, std_dev, total_errs, day_count, modified_at from error_stats where exception = ?`, excp)
	i := new(StatItem)
	var tempDate string
	err := r.Scan(&i.Exception, &i.Mean, &i.Variance, &i.StdDev, &i.TotalErrors, &i.DayCount, &tempDate)
	if err != nil {
		fmt.Printf("Failed mapping stat item: %v\n", err)
	}
	date, err := toDateTime(tempDate)
	i.ModifiedAt = &date
	if err != nil {
		fmt.Printf("Failed parsing date: %v : %v\n", tempDate, err)
	}
	return i
}

func initStats() {
	err := createStatsDBStructure()
	if err != nil {
		fmt.Printf("Failed initializing Stats DB structure: [%v]\n", err)
	}
	err = updateErrorDaySummaries()
	if err != nil {
		fmt.Printf("Failed loading Day Summaries: [%v]\n", err)
	}
	fmt.Println("Day summaries for errors initialized")

	calcStats()
}

type Summary struct {
	Id        int
	Date      time.Time
	Exception string
	Total     int
}

type StatItem struct {
	Exception   string
	Mean        float64
	Variance    int
	StdDev      float64
	TotalErrors int
	DayCount    int
	ModifiedAt  *time.Time
}

func calcStats() {
	summaries := fetchDaySummaries()
	var statMap map[string][]Summary = make(map[string][]Summary)
	for _, s := range summaries {
		if item, ok := statMap[s.Exception]; ok {
			statMap[s.Exception] = append(item, s)
		} else {
			statMap[s.Exception] = append([]Summary{}, s)
		}
	}
	for k, v := range statMap {
		total := calcTotal(v)
		avg := float64(total / len(v))
		variance := calcVariance(v, avg)
		stdDev := math.Sqrt(float64(variance))
		now := time.Now()
		statItem := StatItem{k, avg, variance, stdDev, total, len(v), &now}
		err := insertOrUpdateErrorStat(&statItem)
		if err != nil {
			fmt.Printf("Failed inserting Stat Item for: [%v] : %v\n", k, err)
		} else {
			fmt.Printf("Inserted StatItem for -> %v\n", k)
		}
	}
}

func insertOrUpdateErrorStat(s *StatItem) error {
	var date = s.ModifiedAt.Format(errorlog.DATE_FORMAT)
	_, err := db.Exec(`insert into error_stats(exception, mean, variance, std_dev, total_errs, day_count, modified_at) values (?, ?, ?, ?, ?, ?, ?)`,
		&s.Exception, &s.Mean, &s.Variance, &s.StdDev, &s.TotalErrors, &s.DayCount, &date)
	if err != nil {
		fmt.Println("Assuming insert failed because record already exists trying UPDATE")
		_, err := db.Exec(`UPDATE error_stats SET mean = ?, variance = ?, std_dev = ?, total_errs = ?, day_count = ?, modified_at = ? WHERE exception = ?`,
			&s.Mean, &s.Variance, &s.StdDev, &s.TotalErrors, &s.DayCount, &date, &s.Exception)
		if err != nil {
			fmt.Println("Failed UPDATING DB with [%v] : %v\n", *s, err)
		}
	}
	return err
}

func calcTotal(summaries []Summary) int {
	var total int
	for _, s := range summaries {
		total += s.Total
	}
	return total
}

func calcVariance(summaries []Summary, avg float64) int {
	var variance int
	for _, s := range summaries {
		diff := float64(s.Total) - avg
		variance += int(math.Pow(diff, 2))
	}
	return variance / len(summaries)
}

func fetchDaySummaries() []Summary {
	var summaries []Summary
	rows, err := db.Query("select * from error_day_summary")
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

func toDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, date)
}

func toDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}

func getDaySummaryFor(event *errorlog.ErrorEvent) *Summary {
	s := new(Summary)
	var tempDate string
	/*
		Scan into tempDate string since Scan can't automatically figure out the Date format. So we scan to a string and parse the string with a known date layout
	*/
	err := db.QueryRow("select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events where error_date = DATE(?) group by DATE(error_date), exception",
		event.Timestamp).Scan(&tempDate, &s.Exception, &s.Total)
	if err != nil {
		fmt.Printf("Failed to map Day Summary for [%v] : %v\n", *event, err)
	}
	date, err := toDate(tempDate)
	if err != nil {
		fmt.Printf("Unknown Date format: %v", tempDate)
	}
	s.Date = date
	return s
}

func updateErrorDaySummaries() error {
	_, err := db.Exec(`
		insert or ignore into error_day_summary(error_date, exception, total) select DATE(event_datetime) as error_date, exception, count(exception) as total from error_events group by DATE(error_date), exception
	`)
	return err
}

func createTable(table string, sql string) error {
	var err error
	if hasTable(table) {
		return ErrTableExists
	} else {
		_, err = db.Exec(sql)
		return err
	}
}

func createStatsDBStructure() error {
	var err error
	table := "error_day_summary"
	err = createTable(table, `
	create table error_day_summary(
		id INTEGER not null primary key,
		error_date DATETIME not null,
		exception VARCHAR(255) not null,
		total INTEGER not null,
		unique(error_date, exception)
	)
	`)
	if err != nil {
		fmt.Printf("Error creating table: [%v]\n", table)
	} else {
		fmt.Printf("Created Table: [%v]\n", table)
	}
	table = "error_stats"
	err = createTable(table, `
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
	`)
	if err != nil {
		fmt.Printf("Error creating table: [%v]\n", table)
	} else {
		fmt.Printf("Created Table: [%v]\n", table)
	}
	return err
}

func loadAll(files []string) {
	fmt.Printf("Files: %v\n", files)
	for _, filePath := range files {
		fmt.Printf("Loading File: %v\n", filePath)
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			fmt.Errorf("Error occured while opening '%v' for reading. Error: %v", "simcontrol.log", err)
		}
		scanner := bufio.NewScanner(file)
		readLogFile(scanner)
	}
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "errors.db")
	if err != nil {
		return nil
	}
	table := "error_events"
	err = createTable(table, `create table error_events(
		id INTEGER not null primary key,
		event_datetime DATETIME not null,
		level VARCHAR(10) not null,
		source VARCHAR(30) not null,
		description VARCHAR(255) not null,
		exception VARCHAR(255) not null,
		excp_description VARCHAR(255) not null,
		unique(event_datetime, source, description)
	)
	`)
	if err == ErrTableExists {
		fmt.Printf("Database Initialize Error: [%v]\n", err)
		return nil
	}
	return err
}

func hasTable(name string) bool {
	var table string
	err := db.QueryRow("select name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&table)
	table = strings.Trim(table, " ")
	if err == sql.ErrNoRows || table == "" {
		return false
	} else {
		return true
	}

}

func watchFile(path string, errorProcess chan errorlog.ErrorEvent) {
	t, _ := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
	var event *errorlog.Event = nil
	fmt.Printf("Tailing file: %v\n", path)
	for l := range t.Lines {
		line := l.Text
		fmt.Printf("Tail: %v\n", line)
		e, err := errorlog.ParseLogLine(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		errorEvent, err := processEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			errorProcess <- *errorEvent
			event = nil
		}
	}
}

func processEventIfIsCausedByLine(line string, event *errorlog.Event) (*errorlog.ErrorEvent, error) {
	if errorlog.ContainsCausedBy(line) {
		errorEvent, err := errorlog.ParseCausedBy(line, event)
		if err != nil {
			return nil, err
		} else {
			err = addToDB(errorEvent)
			if err != nil {
				return nil, err
			} else {
				return errorEvent, nil
			}
		}
	}
	return nil, ErrNotCausedByLine
}

func addToDB(errorEvent *errorlog.ErrorEvent) error {
	var count int
	db.QueryRow(`select count(id) from error_events where event_datetime=? AND source=? AND description=? AND exception=? AND excp_description=?`,
		errorEvent.Timestamp, errorEvent.Source, errorEvent.Description, errorEvent.Exception, errorEvent.Description).Scan(&count)
	if count > 0 {
		fmt.Printf("[%v : %v] Already exists!\n", *errorEvent.Timestamp, errorEvent.Exception)
		return nil
	}
	_, err := db.Exec(`insert into error_events(event_datetime, level, source, description, exception, excp_description) 
	values (?, ?, ?, ?, ?, ?)`, errorEvent.Timestamp, string(errorEvent.Level), errorEvent.Source, errorEvent.Description, errorEvent.Exception, errorEvent.Description)
	if err != nil {
		return err
	}
	return nil
}

func readLogFile(scanner *bufio.Scanner) {
	var event *errorlog.Event = nil
	for scanner.Scan() {
		line := scanner.Text()
		e, err := errorlog.ParseLogLine(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		errorEvent, err := processEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			event = nil
		}
	}
}
