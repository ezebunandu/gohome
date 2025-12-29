package weather

import (
	"encoding/json"
	"fmt"
)

type Forecast struct {
    Sunrise int
    Sunset int
}

type OWMResponse struct {
    Sys struct {
        Sunrise int
        Sunset int
    }
}
func ParseResponse(data []byte) (Forecast, error){
    var resp OWMResponse
    err := json.Unmarshal(data, &resp)
    if err != nil {
        return Forecast{}, fmt.Errorf("invalid API response %s: %w", data, err)
    }
    forecast := Forecast{
        Sunrise: resp.Sys.Sunrise,
        Sunset: resp.Sys.Sunset,
    }
    return forecast, nil
}

