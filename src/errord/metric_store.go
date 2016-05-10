package errord

import (
	"database/sql"
)

type MetricStore interface {
	Add(m *MetricEvent) error
}

type metricStore struct {
	db *sql.DB
}

func (store *metricStore) Add(m *MetricEvent) error {
	return nil
}
