// Package usecase は天気予報を取得するユースケースとそれが使うポート（インターフェース）を提供する。
// 次のフェーズで EventBridge Scheduler から起動する取得 Lambda からも同じように呼べる。
package usecase

import (
	"context"
	"log/slog"

	"github.com/posiposi/tempestas/internal/domain"
)

// ForecastCache は取得済みの天気予報を保持するキャッシュのポート。
type ForecastCache interface {
	// GetByLocationAndDate は location の date の予報をキャッシュから返す。
	// キャッシュにない場合は nil, nil を返す。
	GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (*domain.Forecast, error)
	// Save は forecast をキャッシュに保存する。同じ地域と日付の予報がすでにあれば上書きする。
	Save(ctx context.Context, forecast domain.Forecast) error
}

// WeatherRepository は外部 API から天気予報を取得するポート。
type WeatherRepository interface {
	// GetByLocationAndDate は location の date の予報を外部 API から取得する。
	// キャッシュは使わず呼ぶたびに外部 API にアクセスする。
	GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error)
}

// GetDailyForecast は地域と日付を指定して天気予報を返すユースケース。
// 外部 API の呼び出しは料金がかかるためキャッシュにある予報は外部 API を呼ばずに返す。
type GetDailyForecast struct {
	forecastCache     ForecastCache
	weatherRepository WeatherRepository
	logger            *slog.Logger
}

// NewGetDailyForecast は forecastCache と weatherRepository を使い、
// キャッシュへの保存の失敗を logger に出力する GetDailyForecast を返す。
// どの引数にも nil は渡せない。
func NewGetDailyForecast(forecastCache ForecastCache, weatherRepository WeatherRepository, logger *slog.Logger) *GetDailyForecast {
	return &GetDailyForecast{forecastCache: forecastCache, weatherRepository: weatherRepository, logger: logger}
}

// Execute は location の date の予報を返す。
// キャッシュにあればそれを返し、weatherRepository は呼ばない。
// なければ weatherRepository から取得し、キャッシュに保存してから返す。
// キャッシュの読み込みに失敗した場合は外部 API の呼び出しが増えるのを避けるため weatherRepository は呼ばずにそのエラーを返す。
// weatherRepository が返したエラーはそのまま返す。
// キャッシュへの保存に失敗した場合は予報は取得できているのでエラーにせずログを出して予報を返す。
func (u *GetDailyForecast) Execute(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	cached, err := u.forecastCache.GetByLocationAndDate(ctx, location, date)
	if err != nil {
		return domain.Forecast{}, err
	}
	if cached != nil {
		return *cached, nil
	}
	forecast, err := u.weatherRepository.GetByLocationAndDate(ctx, location, date)
	if err != nil {
		return domain.Forecast{}, err
	}
	if err := u.forecastCache.Save(ctx, forecast); err != nil {
		u.logger.ErrorContext(ctx, "failed to save forecast to cache",
			"location", location.Code(), "date", date.String(), "error", err)
	}
	return forecast, nil
}
