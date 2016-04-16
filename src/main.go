package main

import (
	"errord"
	"errors"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"path/filepath"
)

var store errord.Store

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
	store = errord.NewStore()
	errs := store.Init()
	if len(errs) > 0 {
		log.Printf("There were problems initializing the database: [%v]\n", errs)
	} else {
		log.Println("Database initiliazed")
	}
	logParser := errord.NewLogFileParser(store.Errors())
	loadAll(logParser, findAllFilesToParse(oldLogsPath))
	statEngine := errord.NewStatEngine(store)
	statEngine.Init()
	log.Printf("Stat Engine initialized")
	notifier := errord.NewConsoleNotifier(store.Notifications())
	log.Printf("Watching %v", tailPath)
	eventChan := logParser.Watch(tailPath)
	log.Printf("Stat Engine listening for events from watched file")
	go statEngine.ListenOn(eventChan, notifier)
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

func loadAll(parser errord.ErrorParser, files []string) {
	if len(files) == 0 {
		log.Printf("Empty list of files received. Not loading any files")
	}
	for _, filePath := range files {
		log.Printf("Loading File: %v\n", filePath)
		parseStats := parser.Parse(filePath)
		log.Printf("File: %v Stats -> %v", filePath, parseStats)
	}
}
