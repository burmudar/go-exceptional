package errord

import (
	"log"
	"math"
	"time"
)

type Summary struct {
	Exception    string
	StartDate    time.Time
	EndDate      time.Time
	DaySummaries []*DaySummary
	Total        int
}

func (s Summary) DaysInPeriod() int {
	start := s.StartDate.Round(24 * time.Hour)
	end := s.EndDate.Round(24 * time.Hour)
	hours := end.Sub(start).Hours()
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
	Variance    float64
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
	updateStats()
	getStat(event *ErrorEvent) *StatItem
	Listen(n Notifier) chan ErrorEvent
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
	e.updateStats()
}

func (e *statEngine) updateStats() {
	err := e.store.UpdateDaySummaries()
	if err == nil {
		log.Println("Day summaries for errors updated")
	} else {
		log.Printf("Failed updating day summaires: %v\n", err)
	}
	summaries := e.store.FetchSummaries()
	log.Printf("Got %v Summaires. Calculating status on them\n", len(summaries))
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
	avg := s.calcAvg()
	variance := s.calcVariance(avg)
	stdDev := s.calcStdDev(variance)
	now := time.Now()
	statItem := StatItem{s.Exception, avg, variance, stdDev, s.Total, s.DaysInPeriod(), &now}
	return &statItem
}

func (e *statEngine) getStat(event *ErrorEvent) *StatItem {
	return e.store.GetStatItem(event.Exception)
}

func (e *statEngine) Listen(n Notifier) chan ErrorEvent {
	eventBus := make(chan ErrorEvent)
	log.Printf("Creating StatCache")
	cache := createStatCache(e)
	log.Printf("Reading from EventBus")
	go func() {
		for event := range eventBus {
			now := time.Now()
			if cache.shouldReset(&now) {
				cache.reset()
			}
			log.Printf("Processing: %v - %v\n", event.Timestamp, event.Exception)
			log.Printf("Retrieving StatItem for: %v - %v\n", event.Timestamp, event.Exception)
			var statItem *StatItem = cache.get(&event)
			log.Printf("Got: %v\n", statItem)
			if statItem == nil {
				log.Printf("No Stat Item. Exception is propbably new. Notifying of: %v\n", event.Exception)
				notification := &ErrorNotification{}
				notification.ErrorEvent = &event
				n.Fire(notification)
			} else {
				log.Printf("Retrieving DaySummary for: %v - %v\n", event.Timestamp, event.Exception)
				var sum *DaySummary = e.store.GetDaySummary(&event)
				log.Printf("DaySummary: %v - %v [%v]\n", sum.Date, sum.Exception, sum.Total)
				log.Printf("Checking if [%v] exceeds StdMax [%v] ...", sum.Total, statItem.StdDevMax())
				if e.dayTotalExceedsStatLimit(statItem, sum) {
					log.Printf("[%v] exceeds StdMax ... Fire Notification!", event.Exception)
					n.Fire(&ErrorNotification{&event, sum, statItem})
				}
			}

		}
	}()
	return eventBus
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
	c.engine.updateStats()
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
		log.Printf("Adding [%v] Summary with [%v] Day Summaries to Map", s.Exception, len(s.DaySummaries))
		if _, ok := statMap[s.Exception]; !ok {
			statMap[s.Exception] = s
		}
	}
	return statMap
}

func (s Summary) calcStdDev(variance float64) float64 {
	return math.Sqrt(float64(variance))
}

func (s Summary) calcAvg() float64 {
	count := s.DaysInPeriod()
	if count == 0 {
		return 0
	}
	return float64(s.Total / count)
}

func (s Summary) calcVariance(avg float64) float64 {
	var variance float64
	days := s.DaysInPeriod()
	if days == 0 {
		return 0
	}
	for _, day := range s.DaySummaries {
		diff := float64(day.Total) - avg
		variance += math.Pow(diff, 2)
	}
	return variance / float64(days)
}
