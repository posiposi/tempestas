package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/posiposi/tempestas/internal/domain"
)

// JST では 2026-11-17 21:00。
var forecastTestNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

func mustTemperature(t *testing.T, celsius float64) domain.Temperature {
	t.Helper()
	temp, err := domain.NewTemperature(celsius)
	if err != nil {
		t.Fatalf("NewTemperature(%v) returned error: %v", celsius, err)
	}
	return temp
}

func TestNewForecast(t *testing.T) {
	loc, err := domain.NewLocation("tokyo")
	if err != nil {
		t.Fatalf("NewLocation returned error: %v", err)
	}
	date, err := domain.NewForecastDate("2026-11-20", forecastTestNow)
	if err != nil {
		t.Fatalf("NewForecastDate returned error: %v", err)
	}
	precip, err := domain.NewPrecipitationProbability(10)
	if err != nil {
		t.Fatalf("NewPrecipitationProbability returned error: %v", err)
	}

	tests := []struct {
		name string
		max  float64
		min  float64
	}{
		{name: "max_above_min", max: 18.2, min: 9.5},
		{name: "max_equals_min", max: 12.0, min: 12.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxTemp := mustTemperature(t, tt.max)
			minTemp := mustTemperature(t, tt.min)

			f, err := domain.NewForecast(loc, date, domain.WeatherClear, maxTemp, minTemp, precip)
			if err != nil {
				t.Fatalf("NewForecast returned error: %v", err)
			}
			if got := f.Location(); got != loc {
				t.Errorf("Location() = %v, want %v", got, loc)
			}
			if got := f.Date(); got != date {
				t.Errorf("Date() = %v, want %v", got, date)
			}
			if got := f.Weather(); got != domain.WeatherClear {
				t.Errorf("Weather() = %v, want %v", got, domain.WeatherClear)
			}
			if got := f.MaxTemperature(); got != maxTemp {
				t.Errorf("MaxTemperature() = %v, want %v", got, maxTemp)
			}
			if got := f.MinTemperature(); got != minTemp {
				t.Errorf("MinTemperature() = %v, want %v", got, minTemp)
			}
			if got := f.PrecipitationProbability(); got != precip {
				t.Errorf("PrecipitationProbability() = %v, want %v", got, precip)
			}
		})
	}
}

func TestNewForecast_Invalid(t *testing.T) {
	loc, err := domain.NewLocation("tokyo")
	if err != nil {
		t.Fatalf("NewLocation returned error: %v", err)
	}
	date := domain.TodayForecastDate(forecastTestNow)
	precip, err := domain.NewPrecipitationProbability(10)
	if err != nil {
		t.Fatalf("NewPrecipitationProbability returned error: %v", err)
	}

	tests := []struct {
		name string
		max  float64
		min  float64
	}{
		{name: "最高気温が最低気温より低い", max: 9.5, min: 18.2},
		{name: "わずかに逆転", max: 9.4, min: 9.5},
		{name: "氷点下で逆転", max: -5.0, min: -1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewForecast(loc, date, domain.WeatherClear, mustTemperature(t, tt.max), mustTemperature(t, tt.min), precip)
			if !errors.Is(err, domain.ErrInvertedTemperatureRange) {
				t.Errorf("NewForecast(max=%v, min=%v) error = %v, want %v", tt.max, tt.min, err, domain.ErrInvertedTemperatureRange)
			}
		})
	}
}
