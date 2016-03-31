package errorwatch

import "time"

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
