// Package stats provides ...
package errorwatch

import (
	"log"
	"math"
	"time"
)

type Summary struct {
	Exception    string
	FirstSeen    time.Time
	DaySummaries []*DaySummary
	Total        int
}

func (s Summary) DaysFromFirstSeen(date time.Time) int {
	log.Printf("%v", s.FirstSeen)
	start := s.FirstSeen.Round(24 * time.Hour)
	log.Printf("%v", start)
	log.Printf("%v", date)
	date = date.Round(24 * time.Hour)
	log.Printf("%v", date)
	hours := date.Sub(start).Hours()

	return int(math.Ceil(hours / 24))
}

type DaySummary struct {
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

func (s *StatItem) StdDevMax() int {
	return int(s.StdDev + s.Mean)
}

type StatEngine interface {
	Init()
	calcStats(summary []Summary) []*StatItem
	UpdateStats()
	getStat(event *ErrorEvent) *StatItem
	ListenOn(eventBus chan ErrorEvent, n Notifier)
}

type statEngine struct {
	store StatStore
}

func NewStatEngine(s Store) StatEngine {
	e := new(statEngine)
	e.store = s.Stats()
	return e
}

func (e *statEngine) Init() {
	e.UpdateStats()
}

func (e *statEngine) UpdateStats() {
	err := e.store.UpdateDaySummaries()
	if err == nil {
		log.Println("Day summaries for errors updated")
	}
	summaries := e.store.FetchSummaries()
	stats := e.calcStats(summaries)
	for _, stat := range stats {
		err := e.store.InsertOrUpdateStatItem(stat)
		if err != nil {
			log.Printf("Failed inserting Stat: [%v] : %v\n", stat, err)
		} else {
			log.Printf("Inserted StatItem -> %v\n", stat)
		}
	}
}

func (e *statEngine) calcStats(summaries []Summary) []*StatItem {
	log.Printf("Crunching Day summaries to update Stat Items for all exceptions\n")
	statMap := createMapWithSummaries(summaries)
	stats := make([]*StatItem, 0)
	for excp, summaries := range statMap {
		log.Printf("Calculating stats for [%v]\n", excp)
		statItem := createStatItem(summaries)
		stats = append(stats, statItem)
	}
	log.Printf("Crunching COMPLETE!")
	return stats
}

func createStatItem(s Summary) *StatItem {
	avg := calcAvg(s)
	variance := calcVariance(s, avg)
	stdDev := calcStdDev(variance)
	now := time.Now()
	statItem := StatItem{s.Exception, avg, variance, stdDev, s.Total, len(s.DaySummaries), &now}
	return &statItem
}

func (e *statEngine) getStat(event *ErrorEvent) *StatItem {
	return e.store.GetStatItem(event.Exception)
}

func (e *statEngine) ListenOn(eventBus chan ErrorEvent, n Notifier) {
	cache := createStatCache(e)
	for event := range eventBus {
		now := time.Now()
		if cache.shouldReset(&now) {
			cache.reset()
		}
		log.Printf("Processing: %v\n", event)
		var statItem *StatItem = cache.get(&event)
		if statItem == nil {
			log.Printf("No Stat Item. Exception is propbably new. Notifying of: %v\n", event.Exception)
			notification := &ErrorNotification{}
			notification.ErrorEvent = &event
			n.Fire(notification)
		} else {
			var sum *DaySummary = e.store.GetDaySummary(&event)
			if e.dayTotalExceedsStatLimit(statItem, sum) {
				n.Fire(&ErrorNotification{&event, sum, statItem})
			}
		}

	}
}

func (e *statEngine) dayTotalExceedsStatLimit(stat *StatItem, sum *DaySummary) bool {
	if sum == nil {
		return false
	} else {
		return sum.Total >= stat.StdDevMax()
	}
}

type statCache struct {
	start  *time.Time
	cache  map[string]*StatItem
	engine StatEngine
}

func createStatCache(engine StatEngine) *statCache {
	c := new(statCache)
	c.start, c.cache = initStartAndMap()
	c.engine = engine
	return c
}

func (c *statCache) shouldReset(t *time.Time) bool {
	return t.Sub(*c.start).Hours() >= 24
}

func (c *statCache) reset() {
	c.start, c.cache = initStartAndMap()
	c.engine.UpdateStats()
}

func initStartAndMap() (*time.Time, map[string]*StatItem) {
	m := make(map[string]*StatItem)
	return createTimeAtStartOfToday(), m
}

func createTimeAtStartOfToday() *time.Time {
	now := time.Now()
	var start time.Time = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &start
}

func (c *statCache) get(event *ErrorEvent) *StatItem {
	var item *StatItem
	var ok bool
	if item, ok = c.cache[event.Exception]; !ok {
		item = c.engine.getStat(event)
		c.cache[event.Exception] = item
	}
	return item
}

func createMapWithSummaries(summaries []Summary) map[string]Summary {
	var statMap map[string]Summary = make(map[string]Summary)
	for _, s := range summaries {
		if _, ok := statMap[s.Exception]; !ok {
			statMap[s.Exception] = s
		}
	}
	return statMap
}

func calcStdDev(variance int) float64 {
	return math.Sqrt(float64(variance))
}

func calcAvg(s Summary) float64 {
	now := time.Now()
	count := s.DaysFromFirstSeen(now)
	if count == 0 {
		return 0
	}
	return float64(s.Total / count)
}

func calcVariance(s Summary, avg float64) int {
	var variance int
	for _, day := range s.DaySummaries {
		diff := float64(day.Total) - avg
		variance += int(math.Pow(diff, 2))
	}
	return variance / len(s.DaySummaries)
}
