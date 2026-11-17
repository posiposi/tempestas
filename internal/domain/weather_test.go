package domain_test

import (
	"errors"
	"testing"

	"github.com/posiposi/tempestas/internal/domain"
)

func TestNewWeather(t *testing.T) {
	tests := []struct {
		s    string
		want domain.Weather
	}{
		{s: "clear", want: domain.WeatherClear()},
		{s: "cloudy", want: domain.WeatherCloudy()},
		{s: "fog", want: domain.WeatherFog()},
		{s: "rain", want: domain.WeatherRain()},
		{s: "snow", want: domain.WeatherSnow()},
		{s: "thunderstorm", want: domain.WeatherThunderstorm()},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			w, err := domain.NewWeather(tt.s)
			if err != nil {
				t.Fatalf("NewWeather(%q) returned error: %v", tt.s, err)
			}
			if w != tt.want {
				t.Errorf("NewWeather(%q) = %v, want %v", tt.s, w, tt.want)
			}
		})
	}
}

func TestNewWeather_Invalid(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{name: "空文字", s: ""},
		{name: "先頭が大文字", s: "Clear"},
		{name: "すべて大文字", s: "CLEAR"},
		{name: "前後に空白", s: " clear "},
		{name: "未定義の語", s: "sunny"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := domain.NewWeather(tt.s)
			if !errors.Is(err, domain.ErrInvalidWeather) {
				t.Errorf("NewWeather(%q) error = %v, want %v", tt.s, err, domain.ErrInvalidWeather)
			}
			if w != (domain.Weather{}) {
				t.Errorf("NewWeather(%q) = %v, want zero value", tt.s, w)
			}
		})
	}
}

func TestWeather_String(t *testing.T) {
	tests := []struct {
		name string
		w    domain.Weather
		want string
	}{
		{name: "clear", w: domain.WeatherClear(), want: "clear"},
		{name: "cloudy", w: domain.WeatherCloudy(), want: "cloudy"},
		{name: "fog", w: domain.WeatherFog(), want: "fog"},
		{name: "rain", w: domain.WeatherRain(), want: "rain"},
		{name: "snow", w: domain.WeatherSnow(), want: "snow"},
		{name: "thunderstorm", w: domain.WeatherThunderstorm(), want: "thunderstorm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// String で得た文字列を NewWeather に渡すと元の値に戻ることを確かめる（DynamoDB への保存と復元を想定）。
func TestWeather_RoundTrip(t *testing.T) {
	all := []domain.Weather{
		domain.WeatherClear(),
		domain.WeatherCloudy(),
		domain.WeatherFog(),
		domain.WeatherRain(),
		domain.WeatherSnow(),
		domain.WeatherThunderstorm(),
	}
	for _, w := range all {
		t.Run(w.String(), func(t *testing.T) {
			got, err := domain.NewWeather(w.String())
			if err != nil {
				t.Fatalf("NewWeather(%q) returned error: %v", w.String(), err)
			}
			if got != w {
				t.Errorf("NewWeather(%q) = %v, want %v", w.String(), got, w)
			}
		})
	}
}
