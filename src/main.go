package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/hpcloud/tail"
	_ "github.com/mattn/go-sqlite3"
	"logevent"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var db *sql.DB

type CausedBy struct {
	Exception   string
	Description string
}

func main() {

	files := []string{}
	filepath.Walk("logs", func(p string, i os.FileInfo, err error) error {
		if path.Ext(p) == ".log" {
			files = append(files, p)
		}
		return nil
	})
	startReading(files)
}

func startReading(files []string) {
	var err error
	db, err = sql.Open("sqlite3", "errors.db")
	if err != nil {
		fmt.Println("Failed to create database")
		return
	}
	defer db.Close()
	if hasTable("error_events") {
		fmt.Println("DB already has table. Not creating table")
	} else {
		err = initDB()
		if err != nil {
			fmt.Printf("Failed to create required tables: %v\n", err)
			return
		}
	}
	fmt.Printf("Files: %v\n", files)
	for _, filePath := range files {
		fmt.Printf("Loading File: %v\n", filePath)
		readLogFileUsingScanner(filePath)
	}
}

func initDB() error {
	_, err := db.Exec(`
	create table error_events(
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
	fmt.Println("table created")
	return err
}

func hasTable(name string) bool {
	var table string
	err := db.QueryRow("select name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&table)
	table = strings.Trim(table, " ")
	fmt.Printf("Table: %v\n", table)
	if err == sql.ErrNoRows || table == "" {
		return false
	} else {
		return true
	}

}

func readLogFileUsingTail() {
	t, _ := tail.TailFile("simcontrol.log", tail.Config{Follow: true, ReOpen: true})
	var event *logevent.LogEvent = nil
	for l := range t.Lines {
		line := l.Text
		e, err := logevent.Parse(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		if event != nil && strings.HasPrefix(line, "Caused by:") {
			causedBy := createCausedByFromLine(&line)
			err = addToDB(event, causedBy)
			if err != nil {
				fmt.Printf("Failed inserting error event in DB: %v\n", err)
			}
			event = nil
		}
	}
}

func createCausedByFromLine(line *string) *CausedBy {
	parts := strings.Split(*line, ":")
	var causedBy *CausedBy = new(CausedBy)
	causedBy.Exception = parts[1]
	if len(parts) == 3 {
		causedBy.Description = parts[2]
	} else {
		//some exceptions don't have causes ex. "Caused By: java.xml.UnmarshallException"
		causedBy.Description = ""
	}
	return causedBy
}

func addToDB(event *logevent.LogEvent, causedBy *CausedBy) error {
	var count int
	db.QueryRow(`select count(id) from error_events where event_datetime=? AND source=? AND description=? AND exception=? AND excp_description=?`,
		event.Timestamp, event.Source, event.Description, causedBy.Exception, causedBy.Description).Scan(&count)
	if count > 0 {
		fmt.Printf("Count %v : Already Contains %v : %v", count, event, causedBy)
		return nil
	}
	r, err := db.Exec(`insert into error_events(event_datetime, level, source, description, exception, excp_description) 
	values (?, ?, ?, ?, ?, ?)`, event.Timestamp, string(event.Level), event.Source, event.Description, causedBy.Exception, causedBy.Description)
	if err != nil {
		return err
	}
	fmt.Printf("Rows Affected: %d\n", r.RowsAffected)
	fmt.Printf("Last ID: %d\n", r.LastInsertId)
	return nil
}

func readLogFileUsingScanner(logFilePath string) {
	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Errorf("Error occured while opening '%v' for reading. Error: %v", "simcontrol.log", err)
	}
	scanner := bufio.NewScanner(file)
	var event *logevent.LogEvent = nil
	for scanner.Scan() {
		line := scanner.Text()
		e, err := logevent.Parse(line)
		if err != nil {
			fmt.Errorf("Failed parsing: %v\n", err)
		} else {
			event = e
		}
		if event != nil && strings.HasPrefix(line, "Caused by:") {
			causedBy := createCausedByFromLine(&line)
			err = addToDB(event, causedBy)
			if err != nil {
				fmt.Printf("Failed inserting error event in DB: %v\n", err)
			}
			event = nil
		}
	}
}
