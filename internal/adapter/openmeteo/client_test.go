package openmeteo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/posiposi/tempestas/internal/adapter/openmeteo"
	"github.com/posiposi/tempestas/internal/domain"
	"github.com/posiposi/tempestas/internal/usecase"
)

// JST では 2026-11-23 12:00。
var clientTestNow = time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)

const testTimeout = 2 * time.Second

// okResponse は Open-Meteo の実際の応答と同じ形の、2026-11-25 の1日分の予報。
const okResponse = `{
  "latitude": 35.7,
  "longitude": 139.6875,
  "timezone": "Asia/Tokyo",
  "daily_units": {"time": "iso8601", "weather_code": "wmo code", "temperature_2m_max": "°C", "temperature_2m_min": "°C", "precipitation_probability_max": "%"},
  "daily": {
    "time": ["2026-11-25"],
    "weather_code": [61],
    "temperature_2m_max": [18.2],
    "temperature_2m_min": [9.5],
    "precipitation_probability_max": [40]
  }
}`

func locationAndDateFixture(t *testing.T) (domain.Location, domain.ForecastDate) {
	t.Helper()
	loc, err := domain.NewLocation("tokyo")
	if err != nil {
		t.Fatalf("NewLocation returned error: %v", err)
	}
	date, err := domain.NewForecastDate("2026-11-25", clientTestNow)
	if err != nil {
		t.Fatalf("NewForecastDate returned error: %v", err)
	}
	return loc, date
}

func TestClient_GetByLocationAndDate(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okResponse))
	}))
	t.Cleanup(srv.Close)
	client := openmeteo.NewClient(srv.URL, testTimeout)

	got, err := client.GetByLocationAndDate(context.Background(), loc, date)
	if err != nil {
		t.Fatalf("GetByLocationAndDate returned error: %v", err)
	}

	if gotPath != "/v1/forecast" {
		t.Errorf("request path = %q, want %q", gotPath, "/v1/forecast")
	}
	wantQuery := map[string]string{
		"latitude":   "35.6895",
		"longitude":  "139.6917",
		"daily":      "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max",
		"timezone":   "Asia/Tokyo",
		"start_date": "2026-11-25",
		"end_date":   "2026-11-25",
	}
	for key, want := range wantQuery {
		if g := gotQuery.Get(key); g != want {
			t.Errorf("query %s = %q, want %q", key, g, want)
		}
	}

	if got.Location() != loc {
		t.Errorf("Location() = %v, want %v", got.Location(), loc)
	}
	if got.Date() != date {
		t.Errorf("Date() = %v, want %v", got.Date(), date)
	}
	if got.Weather().String() != "rain" {
		t.Errorf("Weather() = %q, want %q", got.Weather().String(), "rain")
	}
	if got.MaxTemperature().Celsius() != 18.2 {
		t.Errorf("MaxTemperature() = %v, want 18.2", got.MaxTemperature().Celsius())
	}
	if got.MinTemperature().Celsius() != 9.5 {
		t.Errorf("MinTemperature() = %v, want 9.5", got.MinTemperature().Celsius())
	}
	if got.PrecipitationProbability().Percent() != 40 {
		t.Errorf("PrecipitationProbability() = %v, want 40", got.PrecipitationProbability().Percent())
	}
}

func TestClient_GetByLocationAndDate_UnexpectedStatus(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": true, "reason": "internal error"}`))
	}))
	t.Cleanup(srv.Close)
	client := openmeteo.NewClient(srv.URL, testTimeout)

	_, err := client.GetByLocationAndDate(context.Background(), loc, date)
	if !errors.Is(err, usecase.ErrFetchFailed) {
		t.Errorf("GetByLocationAndDate() error = %v, want %v", err, usecase.ErrFetchFailed)
	}
}

func TestClient_GetByLocationAndDate_ConnectionFailed(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	srv := httptest.NewServer(http.NotFoundHandler())
	baseURL := srv.URL
	srv.Close()
	client := openmeteo.NewClient(baseURL, testTimeout)

	_, err := client.GetByLocationAndDate(context.Background(), loc, date)
	if !errors.Is(err, usecase.ErrFetchFailed) {
		t.Errorf("GetByLocationAndDate() error = %v, want %v", err, usecase.ErrFetchFailed)
	}
}

func TestClient_GetByLocationAndDate_Timeout(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)
	client := openmeteo.NewClient(srv.URL, 50*time.Millisecond)

	start := time.Now()
	_, err := client.GetByLocationAndDate(context.Background(), loc, date)
	elapsed := time.Since(start)

	if !errors.Is(err, usecase.ErrFetchFailed) {
		t.Errorf("GetByLocationAndDate() error = %v, want %v", err, usecase.ErrFetchFailed)
	}
	if elapsed > time.Second {
		t.Errorf("GetByLocationAndDate took %v, want it to give up after the timeout", elapsed)
	}
}
