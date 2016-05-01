package main

import (
	"encoding/json"
	"errord"
	"errors"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
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
var emailConfigPath = ""

type EmailConfig struct {
	Host string
	From string
	Pass string
	To   string
}

func (c EmailConfig) isEmpty() bool {
	return c.Host == "" && c.From == "" && c.Pass == "" && c.To == ""
}

func init() {
	flag.StringVar(&oldLogsPath, "oldLogs", "", "Directory where old .log files are stored and need to be parsed")
	flag.StringVar(&tailPath, "tailFile", "", "location of file to tail and watch")
	flag.StringVar(&emailConfigPath, "emailConfig", "", "Path to email config json. If empty, notifications are written to stdout")
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
	notifier := createNotifier(emailConfigPath, store.Notifications())
	logParser := errord.NewLogFileParser(store.Errors())
	log.Printf("Watching %v", tailPath)
	eventBus := logParser.Watch(tailPath)
	log.Printf("Stat Engine listening for events from event bus")
	statEngine.Listen(eventBus, notifier)
}

func readEmailConfig(path string) EmailConfig {
	var config EmailConfig
	if path == "" {
		return config
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Failed reading file: %v - %v", path, err)
	} else {
		json.Unmarshal(content, &config)
	}
	return config
}

func createNotifier(emailConfigPath string, store errord.NotifyStore) errord.Notifier {
	c := readEmailConfig(emailConfigPath)
	if c.isEmpty() {
		log.Printf("Email Config is empty. Creating Console Notifier")
		return errord.NewConsoleNotifier(store)
	}
	log.Printf("Creating Email Config notifier")
	return errord.NewEmailNotifier(c.Host, c.From, c.Pass, c.To, store)
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
