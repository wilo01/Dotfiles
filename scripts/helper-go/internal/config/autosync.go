package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Future: Work hours and timezone could be configurable via config file
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

// IsWorkDay returns true for Monday through Friday
func IsWorkDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

// ParseTimeString parses "HH:MM" format into hour and minute components
func ParseTimeString(timeStr string) (hour, minute int, err error) {
	n, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil || n != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s (expected HH:MM)", timeStr)
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid time values: %s", timeStr)
	}
	return hour, minute, nil
}

// NextScheduledTime calculates the next run time for a given slot.
// It finds the next occurrence of the slot time, skipping weekends.
// Returns the target time, slot index, and any error.
func NextScheduledTime(schedule ScheduleConfig, targetSlot string) (time.Time, int, error) {
	if len(schedule.Slots) == 0 {
		return time.Time{}, -1, fmt.Errorf("no schedule slots configured")
	}

	// Load timezone -- fail explicitly instead of silently falling back to UTC
	loc, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		return time.Time{}, -1, fmt.Errorf("invalid timezone %q: %w (check preferences.schedule.timezone in config)", schedule.Timezone, err)
	}
	now := time.Now().In(loc)

	// Find matching slot(s)
	var candidates []struct {
		slot  ScheduleSlot
		index int
		next  time.Time
	}

	for i, slot := range schedule.Slots {
		// Filter by target slot if specified
		if targetSlot != "" && slot.Name != targetSlot {
			continue
		}

		hour, minute, err := ParseTimeString(slot.Time)
		if err != nil {
			// If targeting this specific slot, return the error instead of silently skipping
			if targetSlot != "" && slot.Name == targetSlot {
				return time.Time{}, -1, fmt.Errorf("slot %q has invalid time: %w", slot.Name, err)
			}
			continue
		}

		// Calculate next occurrence of this slot
		next := calculateNextOccurrence(now, hour, minute, loc)

		candidates = append(candidates, struct {
			slot  ScheduleSlot
			index int
			next  time.Time
		}{slot, i, next})
	}

	if len(candidates) == 0 {
		if targetSlot != "" {
			return time.Time{}, -1, fmt.Errorf("slot '%s' not found in configuration", targetSlot)
		}
		return time.Time{}, -1, fmt.Errorf("no valid slots found")
	}

	// Sort by next time and return the earliest
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].next.Before(candidates[j].next)
	})

	return candidates[0].next, candidates[0].index, nil
}

// calculateNextOccurrence finds the next occurrence of the given time, skipping weekends
func calculateNextOccurrence(now time.Time, hour, minute int, loc *time.Location) time.Time {
	// Try today first
	candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)

	// If that time has passed today, move to tomorrow
	if !candidate.After(now) {
		candidate = candidate.AddDate(0, 0, 1)
	}

	// Skip weekends
	for !IsWorkDay(candidate) {
		candidate = candidate.AddDate(0, 0, 1)
	}

	return candidate
}

// ValidateSlotWithinWorkHours checks if a slot time falls within work hours
func ValidateSlotWithinWorkHours(slotTime, workStart, workEnd string) error {
	slotHour, slotMin, err := ParseTimeString(slotTime)
	if err != nil {
		return err
	}

	startHour, startMin, err := ParseTimeString(workStart)
	if err != nil {
		return fmt.Errorf("invalid work_hours_start: %w", err)
	}

	endHour, endMin, err := ParseTimeString(workEnd)
	if err != nil {
		return fmt.Errorf("invalid work_hours_end: %w", err)
	}

	slotMins := slotHour*60 + slotMin
	startMins := startHour*60 + startMin
	endMins := endHour*60 + endMin

	if slotMins < startMins || slotMins >= endMins {
		return fmt.Errorf("slot time %s is outside work hours (%s - %s)", slotTime, workStart, workEnd)
	}

	return nil
}

// GetScheduleConfig loads schedule configuration from config file
func GetScheduleConfig() (*ScheduleConfig, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	return &cfg.Preferences.Schedule, nil
}
