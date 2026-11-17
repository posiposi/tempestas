package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/posiposi/tempestas/internal/domain"
)

func TestNewTemperature(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
	}{
		{name: "正の値", celsius: 25.3},
		{name: "負の値", celsius: -12.4},
		{name: "0", celsius: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp, err := domain.NewTemperature(tt.celsius)
			if err != nil {
				t.Fatalf("NewTemperature(%v) returned error: %v", tt.celsius, err)
			}
			if got := temp.Celsius(); got != tt.celsius {
				t.Errorf("Celsius() = %v, want %v", got, tt.celsius)
			}
		})
	}
}

func TestNewTemperature_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
	}{
		{name: "NaN", celsius: math.NaN()},
		{name: "正の無限大", celsius: math.Inf(1)},
		{name: "負の無限大", celsius: math.Inf(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewTemperature(tt.celsius)
			if !errors.Is(err, domain.ErrInvalidTemperature) {
				t.Errorf("NewTemperature(%v) error = %v, want %v", tt.celsius, err, domain.ErrInvalidTemperature)
			}
		})
	}
}
