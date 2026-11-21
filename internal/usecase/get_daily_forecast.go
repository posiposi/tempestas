// Package usecase は天気予報を取得するユースケースと、それが使うポート（インターフェース）を提供する。
// HTTP や Lambda などの入口には依存しないので、API のハンドラからも、
// 次のフェーズで EventBridge Scheduler から起動する取得 Lambda からも同じように呼べる。
package usecase

import (
	"context"

	"github.com/posiposi/tempestas/internal/domain"
)

// WeatherFetcher は外部 API から天気予報を取得するポート。
type WeatherFetcher interface {
	// Fetch は location の date の予報を外部 API から取得する。
	Fetch(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error)
}

// GetDailyForecast は地域と日付を指定して天気予報を返すユースケース。
// 生成は NewGetDailyForecast で行う。
type GetDailyForecast struct {
	fetcher WeatherFetcher
}

// NewGetDailyForecast は fetcher で予報を取得する GetDailyForecast を返す。
func NewGetDailyForecast(fetcher WeatherFetcher) *GetDailyForecast {
	return &GetDailyForecast{fetcher: fetcher}
}

// Execute は location の date の予報を返す。
// fetcher が返したエラーはそのまま返す。
func (u *GetDailyForecast) Execute(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	return u.fetcher.Fetch(ctx, location, date)
}
