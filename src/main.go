package main

import (
	"bufio"
	"errors"
	"errorwatch"
	"flag"
	"github.com/hpcloud/tail"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"path/filepath"
)

var store errorwatch.Store

var ErrNotCausedByLine error = errors.New("Line does not contain 'Caused by'")

var oldLogsPath = ""
var tailPath = ""

func init() {
	flag.StringVar(&oldLogsPath, "oldLogs", "", "Directory where old .log files are stored and need to be parsed")
	flag.StringVar(&tailPath, "tailFile", "", "location of file to tail and watch")
}

func main() {
	flag.Parse()
	log.Println("Starting Error Watcher")
	defer log.Println("Error Watcher exiting")

	if tailPath == "" {
		log.Fatalf("No File given to Tail and watch")
	}
	store = errorwatch.NewStore()
	errs := store.Init()
	if len(errs) > 0 {
		log.Printf("There were problems initializing the database: [%v]\n", errs)
	} else {
		log.Println("Database initiliazed")
	}
	loadAll(findAllFilesToParse(oldLogsPath))
	statEngine := errorwatch.NewStatEngine(store)
	statEngine.Init()
	log.Printf("Stat Engine initialized")
	notifier := errorwatch.NewConsoleNotifier(store.Notifications())
	var eventProcess chan errorwatch.ErrorEvent = make(chan errorwatch.ErrorEvent)
	go statEngine.ListenOn(eventProcess, notifier)
	watchFile(tailPath, eventProcess)
}

func findAllFilesToParse(dir string) []string {
	files := []string{}
	if dir == "" {
		log.Printf("No Old logs directory given to parse. Returning empty array of files to parse")
		return files
	}
	filepath.Walk(dir, func(p string, i os.FileInfo, err error) error {
		if path.Ext(p) == ".log" {
			files = append(files, p)
		}
		return nil
	})
	return files
}
func notify(e *errorwatch.ErrorEvent, stat *errorwatch.StatItem, sum *errorwatch.Summary) {

}

func loadAll(files []string) {
	if len(files) == 0 {
		log.Printf("Empty list of files received. Not loading any files")
	}
	for _, filePath := range files {
		log.Printf("Loading File: %v\n", filePath)
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			log.Printf("Error occured while opening '%v' for reading. Error: %v", "simcontrol.log", err)
			continue
		}
		readLogFile(file)
	}
}

func watchFile(path string, errorProcess chan errorwatch.ErrorEvent) {
	t, _ := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
	var event *errorwatch.Event = nil
	log.Printf("Tailing file: %v\n", path)
	for l := range t.Lines {
		line := l.Text
		e, err := errorwatch.ParseLogLine(line)
		if err != nil {
		} else {
			event = e
		}
		errorEvent, err := addEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			log.Printf("Error Event found in tailed file: %v\n", errorEvent)
			errorProcess <- *errorEvent
			event = nil
		}
	}
}

func addEventIfIsCausedByLine(line string, event *errorwatch.Event) (*errorwatch.ErrorEvent, error) {
	if errorwatch.ContainsCausedBy(line) {
		errorEvent, err := errorwatch.ParseCausedBy(line, event)
		if err != nil {
			return nil, err
		} else {
			err = store.Errors().AddErrorEvent(errorEvent)
			if err != nil {
				return nil, err
			} else {
				return errorEvent, nil
			}
		}
	}
	return nil, ErrNotCausedByLine
}
func readLogFile(file *os.File) {
	scanner := bufio.NewScanner(file)
	var event *errorwatch.Event = nil
	count := 0
	insert := 0
	failed := 0
	for scanner.Scan() {
		line := scanner.Text()
		count++
		e, err := errorwatch.ParseLogLine(line)
		if err != nil {
			failed++
		} else {
			event = e
		}
		errorEvent, err := addEventIfIsCausedByLine(line, event)
		if errorEvent != nil {
			insert++
			event = nil
		}
	}
	log.Printf("Finished parsing [%v]: Lines=%v Failed: %v Inserted Error Events: %v\n", file.Name(), count, failed, insert)
}
