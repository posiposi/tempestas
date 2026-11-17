package domain_test

import (
	"errors"
	"testing"

	"github.com/posiposi/tempestas/internal/domain"
)

func TestNewPrecipitationProbability(t *testing.T) {
	tests := []struct {
		name    string
		percent int
	}{
		{name: "下限", percent: 0},
		{name: "範囲内", percent: 50},
		{name: "上限", percent: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := domain.NewPrecipitationProbability(tt.percent)
			if err != nil {
				t.Fatalf("NewPrecipitationProbability(%d) returned error: %v", tt.percent, err)
			}
			if got := p.Percent(); got != tt.percent {
				t.Errorf("Percent() = %d, want %d", got, tt.percent)
			}
		})
	}
}

func TestNewPrecipitationProbability_OutOfRange(t *testing.T) {
	tests := []struct {
		name    string
		percent int
	}{
		{name: "下限より1小さい", percent: -1},
		{name: "上限より1大きい", percent: 101},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewPrecipitationProbability(tt.percent)
			if !errors.Is(err, domain.ErrPrecipitationProbabilityOutOfRange) {
				t.Errorf("NewPrecipitationProbability(%d) error = %v, want %v", tt.percent, err, domain.ErrPrecipitationProbabilityOutOfRange)
			}
		})
	}
}
