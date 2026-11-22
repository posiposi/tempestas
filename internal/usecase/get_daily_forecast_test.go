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

type fakeWeatherRepository struct {
	forecast    domain.Forecast
	gotLocation domain.Location
	gotDate     domain.ForecastDate
}

func (f *fakeWeatherRepository) GetByLocationAndDate(_ context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	f.gotLocation = location
	f.gotDate = date
	return f.forecast, nil
}

func forecastFixture(t *testing.T, location domain.Location, date domain.ForecastDate) domain.Forecast {
	t.Helper()
	weather, err := domain.NewWeather("rain")
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

func TestGetDailyForecast(t *testing.T) {
	loc, err := domain.NewLocation("tokyo")
	if err != nil {
		t.Fatalf("NewLocation returned error: %v", err)
	}
	date, err := domain.NewForecastDate("2026-11-20", getDailyForecastTestNow)
	if err != nil {
		t.Fatalf("NewForecastDate returned error: %v", err)
	}
	fetched := forecastFixture(t, loc, date)
	weatherRepository := &fakeWeatherRepository{forecast: fetched}
	uc := usecase.NewGetDailyForecast(weatherRepository)

	got, err := uc.Execute(context.Background(), loc, date)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != fetched {
		t.Errorf("Execute() = %v, want %v", got, fetched)
	}
	if weatherRepository.gotLocation != loc || weatherRepository.gotDate != date {
		t.Errorf("GetByLocationAndDate called with (%v, %v), want (%v, %v)", weatherRepository.gotLocation, weatherRepository.gotDate, loc, date)
	}
}
