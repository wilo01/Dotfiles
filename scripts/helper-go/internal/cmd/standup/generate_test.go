package standup

import (
	"testing"
	"time"
)

func TestGetSmartDefaultRange(t *testing.T) {
	tests := []struct {
		name           string
		today          time.Time
		wantStartDay   int
		wantStartMonth time.Month
		wantEndDay     int
		wantEndMonth   time.Month
	}{
		{
			name:           "Monday goes back to Friday",
			today:          time.Date(2025, 12, 29, 0, 0, 0, 0, time.Local), // Monday
			wantStartDay:   26,                                              // Friday
			wantStartMonth: time.December,
			wantEndDay:     28,  // Sunday (yesterday)
			wantEndMonth:   time.December,
		},
		{
			name:           "Tuesday shows last 2 days",
			today:          time.Date(2025, 12, 30, 0, 0, 0, 0, time.Local), // Tuesday
			wantStartDay:   28,                                              // Sunday
			wantStartMonth: time.December,
			wantEndDay:     29, // Monday
			wantEndMonth:   time.December,
		},
		{
			name:           "Thursday (Dec 25) shows last 2 days",
			today:          time.Date(2025, 12, 25, 0, 0, 0, 0, time.Local), // Thursday
			wantStartDay:   23,                                              // Tuesday
			wantStartMonth: time.December,
			wantEndDay:     24, // Wednesday
			wantEndMonth:   time.December,
		},
		{
			name:           "Friday shows last 2 days",
			today:          time.Date(2025, 12, 26, 0, 0, 0, 0, time.Local), // Friday
			wantStartDay:   24,                                              // Wednesday
			wantStartMonth: time.December,
			wantEndDay:     25, // Thursday
			wantEndMonth:   time.December,
		},
		{
			name:           "Saturday shows Thu+Fri",
			today:          time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local), // Saturday
			wantStartDay:   25,                                              // Thursday
			wantStartMonth: time.December,
			wantEndDay:     26, // Friday
			wantEndMonth:   time.December,
		},
		{
			name:           "Sunday shows Thu+Fri",
			today:          time.Date(2025, 12, 28, 0, 0, 0, 0, time.Local), // Sunday
			wantStartDay:   25,                                              // Thursday
			wantStartMonth: time.December,
			wantEndDay:     26, // Friday
			wantEndMonth:   time.December,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getSmartDefaultRange(tt.today)

			if start.Day() != tt.wantStartDay || start.Month() != tt.wantStartMonth {
				t.Errorf("start = %v (%s), want day %d %s",
					start, start.Weekday(),
					tt.wantStartDay, tt.wantStartMonth)
			}
			if end.Day() != tt.wantEndDay || end.Month() != tt.wantEndMonth {
				t.Errorf("end = %v (%s), want day %d %s",
					end, end.Weekday(),
					tt.wantEndDay, tt.wantEndMonth)
			}
		})
	}
}
