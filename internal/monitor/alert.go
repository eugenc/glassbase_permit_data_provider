package monitor

import (
	"encoding/json"
	"log"
	"time"
)

type Alert struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	CountyID  string    `json:"county_id,omitempty"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
}

func SendAlert(level, countyID, message, details string) {
	a := Alert{
		Timestamp: time.Now(),
		Level:     level,
		CountyID:  countyID,
		Message:   message,
		Details:   details,
	}
	b, _ := json.Marshal(a)
	log.Printf("ALERT: %s", string(b))
}
