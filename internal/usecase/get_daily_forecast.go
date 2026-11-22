// Package usecase は天気予報を取得するユースケースと、それが使うポート（インターフェース）を提供する。
// HTTP や Lambda などの入口には依存しないので、API のハンドラからも、
// 次のフェーズで EventBridge Scheduler から起動する取得 Lambda からも同じように呼べる。
package usecase

import (
	"context"

	"github.com/posiposi/tempestas/internal/domain"
)

// WeatherRepository は天気予報を取得するポート。
// 予報をキャッシュから返すか外部 API から取得するかは実装で判断する。
type WeatherRepository interface {
	// GetByLocationAndDate は location の date の予報を返す。
	GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error)
}

// GetDailyForecast は地域と日付を指定して天気予報を返すユースケース。
type GetDailyForecast struct {
	weatherRepository WeatherRepository
}

// NewGetDailyForecast は weatherRepository で予報を取得する GetDailyForecast を返す。
func NewGetDailyForecast(weatherRepository WeatherRepository) *GetDailyForecast {
	return &GetDailyForecast{weatherRepository: weatherRepository}
}

// Execute は location の date の予報を返す。
// weatherRepository が返したエラーはそのまま返す。
func (u *GetDailyForecast) Execute(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	return u.weatherRepository.GetByLocationAndDate(ctx, location, date)
}
