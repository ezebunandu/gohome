package weather

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseResponse__CorrectlyParsesJSONData(t *testing.T){
    t.Parallel()
    data, err := os.ReadFile("testdata/weather.json")
    want := Forecast{
        Sunrise: 1766849973,
        Sunset: 1766878523,
    }
    got, err := ParseResponse(data)
    if err != nil {
        t.Fatal(err)
    }
    if !cmp.Equal(want, got) {
        t.Error(cmp.Diff(want, got))
    }
}

func TestParseResponse__ReturnsErrorGivenEmptyData(t *testing.T){
    t.Parallel()
    _, err := ParseResponse([]byte{})
    if err == nil {
        t.Fatal("want error parsing empty response, got nil")
    }
}