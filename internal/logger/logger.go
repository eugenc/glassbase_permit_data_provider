package logger

import (
	"encoding/json"
	"log"
	"time"
)

// Entry is a single JSON log line.
type Entry struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	CountyID string `json:"county_id,omitempty"`
	Message  string `json:"msg"`
	Error    string `json:"error,omitempty"`
}

func Info(countyID, msg string) {
	emit("info", countyID, msg, "")
}

// Error logs an error-level message.
func Error(countyID, msg string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	emit("error", countyID, msg, errStr)
}

func emit(level, countyID, msg, errStr string) {
	b, _ := json.Marshal(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Level:    level,
		CountyID: countyID,
		Message:  msg,
		Error:    errStr,
	})
	log.Println(string(b))
}
