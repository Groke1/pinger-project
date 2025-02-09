package repository

import (
	"backend/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository interface {
	GetPings(ctx context.Context) ([]models.Ping, error)
	AddPings(ctx context.Context, pings []models.Ping) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetPings(ctx context.Context) ([]models.Ping, error) {
	query := `SELECT ip, duration, time_attempt FROM pings
			  ORDER BY ip`

	var pings []models.Ping
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ping models.Ping
		if err := rows.Scan(&ping.IP, &ping.Duration, &ping.TimeAttempt); err != nil {
			return nil, err
		}
		pings = append(pings, ping)
	}
	return pings, nil
}

func (r *repository) AddPings(ctx context.Context, pings []models.Ping) error {
	var values []any
	var stringValues []string
	for ind, ping := range pings {
		values = append(values, ping.IP, ping.Duration, ping.TimeAttempt)
		stringValues = append(stringValues, fmt.Sprintf("($%d, $%d, $%d)",
			ind*3+1, ind*3+2, ind*3+3))
	}

	query := `INSERT INTO pings (ip, duration, time_attempt)
			  VALUES %s ON CONFLICT (ip) DO UPDATE SET
			  duration = EXCLUDED.duration,
			  time_attempt = EXCLUDED.time_attempt`
	query = fmt.Sprintf(query, strings.Join(stringValues, ","))

	if len(values) > 0 {
		if _, err := r.db.ExecContext(ctx, query, values...); err != nil {
			return err
		}
	}
	return nil
}
