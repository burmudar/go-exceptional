// Package stats provides ...
package errorwatch

import (
	"log"
	"math"
	"time"
)

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

func (s *StatItem) StdDevMax() int {
	return int(s.StdDev + s.Mean)
}

type StatEngine interface {
	Init() error
	Calc()
	getStat(event *ErrorEvent) *StatItem
}

type statEngine struct {
	store StatStore
}

func NewStatEngine(s Store) StatEngine {
	e := new(statEngine)
	e.store = s.Stats()
	return e
}

func (e *statEngine) Init() error {
	err := e.store.UpdateDaySummaries()
	if err == nil {
		log.Println("Day summaries for errors initialized")
	}
	e.Calc()
	return err
}

func (e *statEngine) Calc() {
	summaries := e.store.FetchDaySummaries()
	statMap := createMapWithSummaries(summaries)
	for k, v := range statMap {
		total := calcTotal(v)
		avg := float64(total / len(v))
		variance := calcVariance(v, avg)
		stdDev := math.Sqrt(float64(variance))
		now := time.Now()
		statItem := StatItem{k, avg, variance, stdDev, total, len(v), &now}
		err := e.store.InsertOrUpdateStatItem(&statItem)
		if err != nil {
			log.Fatalf("Failed inserting Stat Item for: [%v] : %v\n", k, err)
		} else {
			log.Printf("Inserted StatItem for -> %v\n", k)
		}
	}
}

func (e *statEngine) getStat(event *ErrorEvent) *StatItem {
	return e.store.GetStatItem(event.Exception)
}

func (e *statEngine) ListenOn(eventChan chan ErrorEvent, n Notifier) {
	cache := createStatCache(e)
	for event := range eventChan {
		if cache.shouldReset(&event) {
			cache.reset()
		}
		log.Printf("Processing: %v\n", event)
		var statItem *StatItem = cache.get(&event)
		if statItem == nil {
			n.Fire(nil) //notify(&event, nil, nil)
		} else {
			if e.dayTotalExceedsStatLimit(&event, statItem) {
				n.Fire(nil) //(&event, statItem, s)
			}
		}

	}
}

func (e *statEngine) dayTotalExceedsStatLimit(event *ErrorEvent, stat *StatItem) bool {
	var s *Summary = e.store.GetDaySummary(event)
	if s == nil {
		return false
	} else {
		return s.Total >= stat.StdDevMax()
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

func (c *statCache) shouldReset(event *ErrorEvent) bool {
	return c.start.Day()-event.Timestamp.Day() > 0
}

func (c *statCache) reset() {
	c.start, c.cache = initStartAndMap()
	c.engine.Calc()
}

func initStartAndMap() (*time.Time, map[string]*StatItem) {
	var now time.Time = time.Now()
	m := make(map[string]*StatItem)
	return &now, m

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

func createMapWithSummaries(summaries []Summary) map[string][]Summary {
	var statMap map[string][]Summary = make(map[string][]Summary)
	for _, s := range summaries {
		if item, ok := statMap[s.Exception]; ok {
			statMap[s.Exception] = append(item, s)
		} else {
			statMap[s.Exception] = append([]Summary{}, s)
		}
	}
	return statMap
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
