package internal_test

import (
	"help-the-stars/internal"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		date     string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "Valid RFC3339 date",
			date:     "2026-02-14T10:26:00Z",
			expected: time.Date(2026, time.February, 14, 10, 26, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Empty date string",
			date:     "",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "Invalid date format",
			date:     "2026-02-14",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "Malformed date string",
			date:     "not-a-date",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := internal.ParseGhDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !got.Equal(tt.expected) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
