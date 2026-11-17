package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/posiposi/tempestas/internal/domain"
)

// baseNow は JST で 2026-11-17 12:00。
var baseNow = time.Date(2026, 11, 17, 3, 0, 0, 0, time.UTC)

func TestNewForecastDate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		now  time.Time
	}{
		{name: "当日", s: "2026-11-17", now: baseNow},
		{name: "翌日", s: "2026-11-18", now: baseNow},
		{name: "6日後", s: "2026-11-23", now: baseNow},
		{name: "UTC 14:59 は JST 23:59 で当日扱い", s: "2026-11-17", now: time.Date(2026, 11, 17, 14, 59, 59, 0, time.UTC)},
		{name: "UTC 15:00 は JST で翌日になる", s: "2026-11-18", now: time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC)},
		{name: "UTC 15:00 では UTC の7日後が JST の6日後になる", s: "2026-11-24", now: time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC)},
		{name: "月をまたぐ", s: "2026-12-03", now: time.Date(2026, 11, 27, 3, 0, 0, 0, time.UTC)},
		{name: "年をまたぐ", s: "2027-01-02", now: time.Date(2026, 12, 28, 3, 0, 0, 0, time.UTC)},
		{name: "閏日", s: "2028-02-29", now: time.Date(2028, 2, 25, 3, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := domain.NewForecastDate(tt.s, tt.now)
			if err != nil {
				t.Fatalf("NewForecastDate(%q, %v) returned error: %v", tt.s, tt.now, err)
			}
			if got := d.String(); got != tt.s {
				t.Errorf("String() = %q, want %q", got, tt.s)
			}
		})
	}
}

func TestNewForecastDate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		s    string
		now  time.Time
	}{
		{name: "前日", s: "2026-11-16", now: baseNow},
		{name: "7日後", s: "2026-11-24", now: baseNow},
		{name: "UTC 14:59 では JST の7日後", s: "2026-11-24", now: time.Date(2026, 11, 17, 14, 59, 59, 0, time.UTC)},
		{name: "UTC 15:00 では UTC の当日が JST の前日になる", s: "2026-11-17", now: time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC)},
		{name: "空文字", s: "", now: baseNow},
		{name: "日が1桁", s: "2026-11-1", now: baseNow},
		{name: "月と日が1桁", s: "2026-1-5", now: time.Date(2026, 1, 3, 3, 0, 0, 0, time.UTC)},
		{name: "スラッシュ区切り", s: "2026/11/17", now: baseNow},
		{name: "区切りなし", s: "20261117", now: baseNow},
		{name: "存在しない日付", s: "2026-02-30", now: time.Date(2026, 2, 27, 3, 0, 0, 0, time.UTC)},
		{name: "閏年でない年の2月29日", s: "2027-02-29", now: time.Date(2027, 2, 27, 3, 0, 0, 0, time.UTC)},
		{name: "時刻付き", s: "2026-11-17T00:00:00+09:00", now: baseNow},
		{name: "前後に空白", s: " 2026-11-17 ", now: baseNow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewForecastDate(tt.s, tt.now)
			if !errors.Is(err, domain.ErrInvalidDate) {
				t.Errorf("NewForecastDate(%q, %v) error = %v, want %v", tt.s, tt.now, err, domain.ErrInvalidDate)
			}
		})
	}
}

func TestTodayForecastDate(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "UTC 14:59 は JST の当日", now: time.Date(2026, 11, 17, 14, 59, 59, 0, time.UTC), want: "2026-11-17"},
		{name: "UTC 15:00 は JST の翌日", now: time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC), want: "2026-11-18"},
		{name: "JST 以外のタイムゾーンの時刻", now: time.Date(2026, 11, 17, 8, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)), want: "2026-11-18"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := domain.TodayForecastDate(tt.now)
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
			parsed, err := domain.NewForecastDate(tt.want, tt.now)
			if err != nil {
				t.Fatalf("NewForecastDate(%q, %v) returned error: %v", tt.want, tt.now, err)
			}
			if parsed != d {
				t.Errorf("NewForecastDate(%q, %v) = %v, want %v", tt.want, tt.now, parsed, d)
			}
		})
	}
}
