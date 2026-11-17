package domain_test

import (
	"errors"
	"testing"

	"github.com/posiposi/tempestas/internal/domain"
)

func TestNewWeather(t *testing.T) {
	tests := []string{"clear", "cloudy", "fog", "rain", "snow", "thunderstorm"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			w, err := domain.NewWeather(s)
			if err != nil {
				t.Fatalf("NewWeather(%q) returned error: %v", s, err)
			}
			if got := w.String(); got != s {
				t.Errorf("String() = %q, want %q", got, s)
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

// String で得た文字列を NewWeather に渡すと元の値に戻ることを確かめる（DynamoDB への保存と復元を想定）。
func TestWeather_RoundTrip(t *testing.T) {
	tests := []string{"clear", "cloudy", "fog", "rain", "snow", "thunderstorm"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			w, err := domain.NewWeather(s)
			if err != nil {
				t.Fatalf("NewWeather(%q) returned error: %v", s, err)
			}
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
