package postgres

import (
	"database/sql"
	"fmt"
)

type Config struct {
	Host    string
	Port    string
	User    string
	Pass    string
	DBName  string
	SSLMode string
}

func NewPostgres(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.Pass, cfg.SSLMode))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
