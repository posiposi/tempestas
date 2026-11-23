package usecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/posiposi/tempestas/internal/domain"
	"github.com/posiposi/tempestas/internal/usecase"
)

// JST では 2026-11-17 21:00。
var getDailyForecastTestNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

type fakeForecastCache struct {
	forecast    *domain.Forecast
	getErr      error
	saveErr     error
	gotLocation domain.Location
	gotDate     domain.ForecastDate
	saved       *domain.Forecast
}

func (f *fakeForecastCache) GetByLocationAndDate(_ context.Context, location domain.Location, date domain.ForecastDate) (*domain.Forecast, error) {
	f.gotLocation = location
	f.gotDate = date
	return f.forecast, f.getErr
}

func (f *fakeForecastCache) Save(_ context.Context, forecast domain.Forecast) error {
	f.saved = &forecast
	return f.saveErr
}

type fakeWeatherRepository struct {
	forecast    domain.Forecast
	err         error
	called      bool
	gotLocation domain.Location
	gotDate     domain.ForecastDate
}

func (f *fakeWeatherRepository) GetByLocationAndDate(_ context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	f.called = true
	f.gotLocation = location
	f.gotDate = date
	return f.forecast, f.err
}

func forecastFixture(t *testing.T, location domain.Location, date domain.ForecastDate, weatherName string) domain.Forecast {
	t.Helper()
	weather, err := domain.NewWeather(weatherName)
	if err != nil {
		t.Fatalf("NewWeather returned error: %v", err)
	}
	maxTemp, err := domain.NewTemperature(18.2)
	if err != nil {
		t.Fatalf("NewTemperature returned error: %v", err)
	}
	minTemp, err := domain.NewTemperature(9.5)
	if err != nil {
		t.Fatalf("NewTemperature returned error: %v", err)
	}
	precip, err := domain.NewPrecipitationProbability(10)
	if err != nil {
		t.Fatalf("NewPrecipitationProbability returned error: %v", err)
	}
	f, err := domain.NewForecast(location, date, weather, maxTemp, minTemp, precip)
	if err != nil {
		t.Fatalf("NewForecast returned error: %v", err)
	}
	return f
}

func locationAndDateFixture(t *testing.T) (domain.Location, domain.ForecastDate) {
	t.Helper()
	loc, err := domain.NewLocation("tokyo")
	if err != nil {
		t.Fatalf("NewLocation returned error: %v", err)
	}
	date, err := domain.NewForecastDate("2026-11-20", getDailyForecastTestNow)
	if err != nil {
		t.Fatalf("NewForecastDate returned error: %v", err)
	}
	return loc, date
}

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func TestGetDailyForecast_CacheHit(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	cached := forecastFixture(t, loc, date, "clear")
	forecastCache := &fakeForecastCache{forecast: &cached}
	weatherRepository := &fakeWeatherRepository{}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository, newTestLogger(&bytes.Buffer{}))

	got, err := uc.Execute(context.Background(), loc, date)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != cached {
		t.Errorf("Execute() = %v, want %v", got, cached)
	}
	if forecastCache.gotLocation != loc || forecastCache.gotDate != date {
		t.Errorf("ForecastCache.GetByLocationAndDate called with (%v, %v), want (%v, %v)", forecastCache.gotLocation, forecastCache.gotDate, loc, date)
	}
	if weatherRepository.called {
		t.Error("WeatherRepository.GetByLocationAndDate was called, want not called on cache hit")
	}
	if forecastCache.saved != nil {
		t.Errorf("ForecastCache.Save called with %v, want not called on cache hit", *forecastCache.saved)
	}
}

func TestGetDailyForecast_CacheMiss(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	fetched := forecastFixture(t, loc, date, "rain")
	forecastCache := &fakeForecastCache{}
	weatherRepository := &fakeWeatherRepository{forecast: fetched}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository, newTestLogger(&bytes.Buffer{}))

	got, err := uc.Execute(context.Background(), loc, date)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != fetched {
		t.Errorf("Execute() = %v, want %v", got, fetched)
	}
	if forecastCache.gotLocation != loc || forecastCache.gotDate != date {
		t.Errorf("ForecastCache.GetByLocationAndDate called with (%v, %v), want (%v, %v)", forecastCache.gotLocation, forecastCache.gotDate, loc, date)
	}
	if weatherRepository.gotLocation != loc || weatherRepository.gotDate != date {
		t.Errorf("WeatherRepository.GetByLocationAndDate called with (%v, %v), want (%v, %v)", weatherRepository.gotLocation, weatherRepository.gotDate, loc, date)
	}
	if forecastCache.saved == nil {
		t.Fatal("ForecastCache.Save was not called, want called with the fetched forecast")
	}
	if *forecastCache.saved != fetched {
		t.Errorf("ForecastCache.Save called with %v, want %v", *forecastCache.saved, fetched)
	}
}

// 保存に失敗しても取得した予報は返し、失敗はエラーレベルのログに残す。
func TestGetDailyForecast_SaveFailed(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	fetched := forecastFixture(t, loc, date, "snow")
	forecastCache := &fakeForecastCache{saveErr: errors.New("cache unavailable")}
	weatherRepository := &fakeWeatherRepository{forecast: fetched}
	var logs bytes.Buffer
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository, newTestLogger(&logs))

	got, err := uc.Execute(context.Background(), loc, date)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != fetched {
		t.Errorf("Execute() = %v, want %v", got, fetched)
	}

	var record struct {
		Level string `json:"level"`
	}
	if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
		t.Fatalf("want one JSON log record, got %q: %v", logs.String(), err)
	}
	if record.Level != slog.LevelError.String() {
		t.Errorf("log level = %q, want %q", record.Level, slog.LevelError.String())
	}
}

func TestGetDailyForecast_CacheGetFailed(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	getErr := errors.New("cache unavailable")
	forecastCache := &fakeForecastCache{getErr: getErr}
	weatherRepository := &fakeWeatherRepository{forecast: forecastFixture(t, loc, date, "clear")}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository, newTestLogger(&bytes.Buffer{}))

	_, err := uc.Execute(context.Background(), loc, date)
	if !errors.Is(err, getErr) {
		t.Errorf("Execute() error = %v, want %v", err, getErr)
	}
	if errors.Is(err, usecase.ErrFetchFailed) {
		t.Errorf("Execute() error = %v, must not be %v", err, usecase.ErrFetchFailed)
	}
	if weatherRepository.called {
		t.Error("WeatherRepository.GetByLocationAndDate was called, want not called when the cache read fails")
	}
	if forecastCache.saved != nil {
		t.Errorf("ForecastCache.Save called with %v, want not called when the cache read fails", *forecastCache.saved)
	}
}

func TestGetDailyForecast_FetchFailed(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	forecastCache := &fakeForecastCache{}
	weatherRepository := &fakeWeatherRepository{err: fmt.Errorf("%w: status 503", usecase.ErrFetchFailed)}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository, newTestLogger(&bytes.Buffer{}))

	_, err := uc.Execute(context.Background(), loc, date)
	if !errors.Is(err, usecase.ErrFetchFailed) {
		t.Errorf("Execute() error = %v, want %v", err, usecase.ErrFetchFailed)
	}
	if forecastCache.saved != nil {
		t.Errorf("ForecastCache.Save called with %v, want not called when the fetch fails", *forecastCache.saved)
	}
}
