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
	"runtime"
	"sync"
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
	log.Println("Starting ErrorD")
	defer log.Println("ErrorD  exiting")

	runtime.GOMAXPROCS(runtime.NumCPU())

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
	loadAll(store.Errors(), findAllFilesToParse(oldLogsPath))
	statEngine := errord.NewStatEngine(store)
	statEngine.Init()
	log.Printf("Stat Engine initialized")
	notifier := errord.NewConsoleNotifier(store.Notifications())
	log.Printf("Watching %v", tailPath)
	logParser := errord.NewLogFileParser(store.Errors())
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

func loadAll(store errord.ErrorStore, files []string) {
	if len(files) == 0 {
		log.Printf("Empty list of files received. Not loading any files")
	}
	goGroup := new(sync.WaitGroup)
	goGroup.Add(len(files))
	for _, filePath := range files {
		go func(s errord.ErrorStore, path string) {
			parser := errord.NewLogFileParser(s)
			log.Printf("Loading File: %v\n", path)
			parseStats := parser.Parse(path)
			log.Printf("File: %v Stats -> %v", path, parseStats)
			goGroup.Done()
		}(store, filePath)
	}
	goGroup.Wait()
}
