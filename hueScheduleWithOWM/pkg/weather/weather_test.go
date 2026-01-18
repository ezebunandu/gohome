package weather

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseResponse__CorrectlyParsesJSONData(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather.json")
	want := Forecast{
		Sunrise: 1766849973,
		Sunset:  1766878523,
	}
	got, err := ParseResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseResponse__ReturnsErrorGivenEmptyData(t *testing.T) {
	t.Parallel()
	_, err := ParseResponse([]byte{})
	if err == nil {
		t.Fatal("want error parsing empty response, got nil")
	}
}

func TestUnixToCron(t *testing.T) {
	t.Parallel()
	mountainTime, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatalf("failed to load Mountain Time location: %v", err)
	}

	tests := []struct {
		name      string
		timestamp int
		want      string
	}{
		{
			name:      "converts int time to cron syntax",
			timestamp: 1766878523, // from testdata
			want: func() string {
				t := time.Unix(1766878523, 0).In(mountainTime)
				return fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour())
			}(),
		},
		{
			name:      "handles midnight correctly",
			timestamp: int(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()),
			want: func() string {
				// Midnight UTC in Mountain Time (UTC-7 in January, so 5pm previous day)
				t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).In(mountainTime)
				return fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour())
			}(),
		},
		{
			name:      "handles noon correctly",
			timestamp: int(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix()),
			want: func() string {
				// Noon UTC in Mountain Time (UTC-7 in January, so 5am)
				t := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).In(mountainTime)
				return fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour())
			}(),
		},
		{
			name:      "handles arbitrary time correctly",
			timestamp: int(time.Date(2024, 6, 15, 14, 35, 0, 0, time.UTC).Unix()),
			want: func() string {
				// 14:35 UTC in Mountain Time (UTC-6 in June, so 8:35am)
				t := time.Date(2024, 6, 15, 14, 35, 0, 0, time.UTC).In(mountainTime)
				return fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour())
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := UnixToCron(tt.timestamp)
			if got != tt.want {
				t.Errorf("UnixToCron(%d) = %q, want %q", tt.timestamp, got, tt.want)
			}
		})
	}
}
