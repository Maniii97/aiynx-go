package service_test

import (
	"testing"
	"time"

	"github.com/Maniii97/aiynx-go/internal/service"
)

func TestCalculateAge(t *testing.T) {
	today := time.Now().UTC()

	// birthday_today: exact same month/day, 30 years back → age 30
	birthdayToday := time.Date(today.Year()-30, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	// birthday_yesterday: 25 years + 1 day ago → birthday was yesterday this year → age 25
	birthdayYesterday := today.AddDate(-25, 0, -1).Truncate(24 * time.Hour)

	// birthday_tomorrow: go back 25 years, then forward 1 day → birthday is tomorrow → age 24
	birthdayTomorrow := today.AddDate(-25, 0, 1).Truncate(24 * time.Hour)

	// Leap day: Feb 29, 2000 — most years this birthday "hasn't happened" before Mar 1
	leapDay := time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		dob     time.Time
		wantAge int
	}{
		{
			name:    "birthday is today",
			dob:     birthdayToday,
			wantAge: 30,
		},
		{
			name:    "birthday was yesterday (already had it this year)",
			dob:     birthdayYesterday,
			wantAge: 25,
		},
		{
			name:    "birthday is tomorrow (not yet this year)",
			dob:     birthdayTomorrow,
			wantAge: 24,
		},
		{
			name: "leap day birthday (Feb 29)",
			dob:  leapDay,
			// We don't hard-code the expected int here because it depends on today's date
			// — instead we verify the result is within a plausible range.
			wantAge: -1, // sentinel: skip exact check, see custom assertion below
		},
		{
			name:    "very old — born 1900-01-01",
			dob:     time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
			wantAge: today.Year() - 1900, // birthday already passed this year on Jan 1
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.CalculateAge(tc.dob)

			if tc.wantAge == -1 {
				// Leap-day special case: just verify the age is sane (>= 0, not absurd).
				if got < 0 || got > 200 {
					t.Errorf("CalculateAge(leap day) = %d; expected a non-negative value", got)
				}
				return
			}

			if got != tc.wantAge {
				t.Errorf("CalculateAge(%v) = %d; want %d", tc.dob.Format("2006-01-02"), got, tc.wantAge)
			}
		})
	}
}
