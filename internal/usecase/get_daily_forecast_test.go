package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/posiposi/tempestas/internal/domain"
	"github.com/posiposi/tempestas/internal/usecase"
)

// JST では 2026-11-17 21:00。
var getDailyForecastTestNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

type fakeForecastCache struct {
	forecast    *domain.Forecast
	gotLocation domain.Location
	gotDate     domain.ForecastDate
}

func (f *fakeForecastCache) GetByLocationAndDate(_ context.Context, location domain.Location, date domain.ForecastDate) (*domain.Forecast, error) {
	f.gotLocation = location
	f.gotDate = date
	return f.forecast, nil
}

type fakeWeatherRepository struct {
	forecast    domain.Forecast
	called      bool
	gotLocation domain.Location
	gotDate     domain.ForecastDate
}

func (f *fakeWeatherRepository) GetByLocationAndDate(_ context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	f.called = true
	f.gotLocation = location
	f.gotDate = date
	return f.forecast, nil
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

func TestGetDailyForecast_CacheHit(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	cached := forecastFixture(t, loc, date, "clear")
	forecastCache := &fakeForecastCache{forecast: &cached}
	weatherRepository := &fakeWeatherRepository{}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository)

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
}

func TestGetDailyForecast_CacheMiss(t *testing.T) {
	loc, date := locationAndDateFixture(t)
	fetched := forecastFixture(t, loc, date, "rain")
	forecastCache := &fakeForecastCache{}
	weatherRepository := &fakeWeatherRepository{forecast: fetched}
	uc := usecase.NewGetDailyForecast(forecastCache, weatherRepository)

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
}
