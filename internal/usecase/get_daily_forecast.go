// Package usecase は天気予報を取得するユースケースとそれが使うポート（インターフェース）を提供する。
// 次のフェーズで EventBridge Scheduler から起動する取得 Lambda からも同じように呼べる。
package usecase

import (
	"context"

	"github.com/posiposi/tempestas/internal/domain"
)

// ForecastCache は取得済みの天気予報を保持するキャッシュのポート。
type ForecastCache interface {
	// GetByLocationAndDate は location の date の予報をキャッシュから返す。
	// キャッシュにない場合は nil, nil を返す。
	GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (*domain.Forecast, error)
}

// WeatherRepository は外部 API から天気予報を取得するポート。
type WeatherRepository interface {
	// GetByLocationAndDate は location の date の予報を外部 API から取得する。
	GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error)
}

// GetDailyForecast は地域と日付を指定して天気予報を返すユースケース。
// 外部 API の呼び出しは料金がかかるためキャッシュにある予報は外部 API を呼ばずに返す。
type GetDailyForecast struct {
	forecastCache     ForecastCache
	weatherRepository WeatherRepository
}

// NewGetDailyForecast は forecastCache と weatherRepository を使う GetDailyForecast を返す。
// どちらにも nil は渡せない。
func NewGetDailyForecast(forecastCache ForecastCache, weatherRepository WeatherRepository) *GetDailyForecast {
	return &GetDailyForecast{forecastCache: forecastCache, weatherRepository: weatherRepository}
}

// Execute は location の date の予報を返す。
// キャッシュにあればそれを返し、weatherRepository は呼ばない。
// forecastCache と weatherRepository が返したエラーはそのまま返す。
// キャッシュの読み込みに失敗した場合は外部 API の呼び出しが増えるのを避けるためweatherRepository は呼ばない。
func (u *GetDailyForecast) Execute(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	cached, err := u.forecastCache.GetByLocationAndDate(ctx, location, date)
	if err != nil {
		return domain.Forecast{}, err
	}
	if cached != nil {
		return *cached, nil
	}
	return u.weatherRepository.GetByLocationAndDate(ctx, location, date)
}
