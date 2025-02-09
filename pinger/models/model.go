package models

import "time"

type Ping struct {
	IP          string    `json:"ip"`
	Duration    int       `json:"duration"`
	TimeAttempt time.Time `json:"time_attempt"`
}
