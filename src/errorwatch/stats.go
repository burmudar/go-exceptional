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
	calcStats(sumamries []Summary) []*StatItem
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

func (e *statEngine) Init() error {
	err := e.store.UpdateDaySummaries()
	if err == nil {
		log.Println("Day summaries for errors initialized")
	}
	e.UpdateStats()
	return err
}

func (e *statEngine) UpdateStats() {
	summaries := e.store.FetchDaySummaries()
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
	for k, v := range statMap {
		total := calcTotal(v)
		avg := float64(total / len(v))
		variance := calcVariance(v, avg)
		stdDev := math.Sqrt(float64(variance))
		now := time.Now()
		statItem := StatItem{k, avg, variance, stdDev, total, len(v), &now}
		stats = append(stats, &statItem)
	}
	log.Printf("Crunching COMPLETE!")
	return stats
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
			var sum *Summary = e.store.GetDaySummary(&event)
			if e.dayTotalExceedsStatLimit(&event, statItem, sum) {
				n.Fire(&ErrorNotification{&event, sum, statItem})
			}
		}

	}
}

func (e *statEngine) dayTotalExceedsStatLimit(event *ErrorEvent, stat *StatItem, sum *Summary) bool {
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
