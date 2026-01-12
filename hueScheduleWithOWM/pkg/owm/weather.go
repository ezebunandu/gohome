package weather

import (
	"encoding/json"
	"fmt"
	"time"
)

type Forecast struct {
	Sunrise int
	Sunset  int
}

type OWMResponse struct {
	Sys struct {
		Sunrise int
		Sunset  int
	}
}

func ParseResponse(data []byte) (Forecast, error) {
	var resp OWMResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Forecast{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	forecast := Forecast{
		Sunrise: resp.Sys.Sunrise,
		Sunset:  resp.Sys.Sunset,
	}
	return forecast, nil
}

// UnixToCron converts a Unix timestamp to cron syntax format (minute hour * * *)
// The timestamp is converted to Mountain Time for the cron schedule
func UnixToCron(timestamp int) string {
	t := time.Unix(int64(timestamp), 0)
	// Load Mountain Time location (handles both MST and MDT automatically)
	// LoadLocation rarely fails for well-known timezones like "America/Denver"
	mountainTime, _ := time.LoadLocation("America/Denver")
	if mountainTime == nil {
		mountainTime = time.UTC
	}
	t = t.In(mountainTime)
	minute := t.Minute()
	hour := t.Hour()
	return fmt.Sprintf("%d %d * * *", minute, hour)
}
