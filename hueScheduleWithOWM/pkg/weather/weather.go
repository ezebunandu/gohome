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
// The timestamp is in UTC, and Kubernetes CronJobs interpret cron schedules in UTC,
// so we can use the UTC time directly
func UnixToCron(timestamp int) string {
	t := time.Unix(int64(timestamp), 0).UTC()
	minute := t.Minute()
	hour := t.Hour()
	return fmt.Sprintf("%d %d * * *", minute, hour)
}
