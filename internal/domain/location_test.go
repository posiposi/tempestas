package domain_test

import (
	"errors"
	"testing"

	"github.com/posiposi/tempestas/internal/domain"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		code      string
		latitude  float64
		longitude float64
	}{
		{code: "sapporo", latitude: 43.0621, longitude: 141.3544},
		{code: "tokyo", latitude: 35.6895, longitude: 139.6917},
		{code: "nagoya", latitude: 35.1688, longitude: 136.9068},
		{code: "osaka", latitude: 34.6937, longitude: 135.5022},
		{code: "fukuoka", latitude: 33.5902, longitude: 130.4017},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			loc, err := domain.NewLocation(tt.code)
			if err != nil {
				t.Fatalf("NewLocation(%q) returned error: %v", tt.code, err)
			}
			if got := loc.Code(); got != tt.code {
				t.Errorf("Code() = %q, want %q", got, tt.code)
			}
			if got := loc.Latitude(); got != tt.latitude {
				t.Errorf("Latitude() = %v, want %v", got, tt.latitude)
			}
			if got := loc.Longitude(); got != tt.longitude {
				t.Errorf("Longitude() = %v, want %v", got, tt.longitude)
			}
		})
	}
}

func TestNewLocation_Invalid(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "空文字", code: ""},
		{name: "先頭が大文字", code: "Tokyo"},
		{name: "すべて大文字", code: "TOKYO"},
		{name: "前後に空白", code: " tokyo "},
		{name: "未対応の地域", code: "kyoto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewLocation(tt.code)
			if !errors.Is(err, domain.ErrInvalidLocation) {
				t.Errorf("NewLocation(%q) error = %v, want %v", tt.code, err, domain.ErrInvalidLocation)
			}
		})
	}
}
