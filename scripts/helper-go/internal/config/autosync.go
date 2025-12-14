package config

import (
	"os"
	"path/filepath"
	"time"
)

// TODO: MAGIC NUMBER - Work hours and timezone should be configurable via config file
const (
	lastSyncFilename = "last-sync"
	// Work hours: 8:00-16:00 Europe/Dublin
	workHoursStart = 8
	workHoursEnd   = 16
)

// GetLastSyncTime returns the last sync timestamp from state file
func GetLastSyncTime() (time.Time, error) {
	path := filepath.Join(GetConfigDir(), lastSyncFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, string(data))
}

// UpdateLastSyncTime writes current time to state file
func UpdateLastSyncTime() error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, lastSyncFilename)
	timestamp := time.Now().Format(time.RFC3339)
	return os.WriteFile(path, []byte(timestamp), 0644)
}

// SyncedToday returns true if sync already ran today
func SyncedToday() bool {
	lastSync, err := GetLastSyncTime()
	if err != nil {
		return false
	}

	// Load Dublin timezone
	dublin, err := time.LoadLocation("Europe/Dublin")
	if err != nil {
		dublin = time.UTC
	}

	now := time.Now().In(dublin)
	last := lastSync.In(dublin)

	return now.Year() == last.Year() &&
		now.Month() == last.Month() &&
		now.Day() == last.Day()
}

// IsWorkHours returns true if current time is within work hours
// Work hours: 8:00-16:00 Europe/Dublin, Monday-Friday
func IsWorkHours() bool {
	dublin, err := time.LoadLocation("Europe/Dublin")
	if err != nil {
		dublin = time.UTC
	}

	now := time.Now().In(dublin)

	// Check weekday (Monday=1, Sunday=0)
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Check hours
	hour := now.Hour()
	return hour >= workHoursStart && hour < workHoursEnd
}
