// Package openmeteo は Open-Meteo の Forecast API から天気予報を取得する、
// usecase.WeatherRepository の実装を提供する。
// WMO の天気コードからドメインの天気区分への変換など Open-Meteo の表現をドメインの値に変換する処理はこのパッケージに閉じ込める。
package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/posiposi/tempestas/internal/domain"
	"github.com/posiposi/tempestas/internal/usecase"
)

// DefaultBaseURL は Open-Meteo の Forecast API のベース URL。
const DefaultBaseURL = "https://api.open-meteo.com"

// DefaultTimeout は Open-Meteo への1回のリクエスト（接続から応答の読み込みまで）のタイムアウト。
// Lambda のタイムアウト（10秒）より十分短くし、Open-Meteo が応答しないときに Lambda ごと打ち切られる前に
// 502 を返せるようにする。残りの時間はキャッシュの読み書きとレスポンスの返却に使う。
const DefaultTimeout = 5 * time.Second

const dailyVariables = "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max"

const timezone = "Asia/Tokyo"

const maxErrorReasonBytes = 4 << 10

var _ usecase.WeatherRepository = (*Client)(nil)

// Client は Open-Meteo の Forecast API のクライアント。
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient は baseURL（末尾のスラッシュなし。本番では DefaultBaseURL）にリクエストする Client を返す。
// timeout は1回のリクエストのタイムアウトで、0 以下の場合は DefaultTimeout を使う
// （http.Client はタイムアウトが 0 だと無制限に待つため、タイムアウトなしにはできないようにする）。
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

// GetByLocationAndDate は location の予報地点の date の予報を Open-Meteo から取得する。
// 呼ぶたびに Open-Meteo にリクエストする。
// 通信の失敗やタイムアウト、200 以外のステータスや値の欠落、対応表にない天気コード、
// ドメインの検証を通らない値の場合は usecase.ErrFetchFailed を包んだエラーを返す。
// ドメインの検証を通らない場合はドメインのエラーも errors.Is で判別できる。
func (c *Client) GetByLocationAndDate(ctx context.Context, location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	res, err := c.fetchDaily(ctx, location, date)
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("%w: %w", usecase.ErrFetchFailed, err)
	}
	forecast, err := res.toForecast(location, date)
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("%w: %w", usecase.ErrFetchFailed, err)
	}
	return forecast, nil
}

func (c *Client) fetchDaily(ctx context.Context, location domain.Location, date domain.ForecastDate) (*forecastResponse, error) {
	q := url.Values{}
	q.Set("latitude", strconv.FormatFloat(location.Latitude(), 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(location.Longitude(), 'f', -1, 64))
	q.Set("daily", dailyVariables)
	q.Set("timezone", timezone)
	q.Set("start_date", date.String())
	q.Set("end_date", date.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/forecast?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, readErrorReason(resp.Body))
	}
	var res forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &res, nil
}

// readErrorReason は失敗時のレスポンス（{"error": true, "reason": "..."}）から reason を取り出す。
// 取り出せない場合は空文字を返す。
func readErrorReason(body io.Reader) string {
	var e struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(io.LimitReader(body, maxErrorReasonBytes)).Decode(&e); err != nil {
		return ""
	}
	return e.Reason
}

// forecastResponse は Forecast API のレスポンスのうち使う部分。
// 値は日ごとの配列でデータがない日は null になるため欠落を判別できるようポインタで受ける。
type forecastResponse struct {
	Daily struct {
		Time                        []string   `json:"time"`
		WeatherCode                 []*int     `json:"weather_code"`
		Temperature2mMax            []*float64 `json:"temperature_2m_max"`
		Temperature2mMin            []*float64 `json:"temperature_2m_min"`
		PrecipitationProbabilityMax []*int     `json:"precipitation_probability_max"`
	} `json:"daily"`
}

// toForecast は1日分だけを要求したレスポンスを Forecast に変換する。
func (r *forecastResponse) toForecast(location domain.Location, date domain.ForecastDate) (domain.Forecast, error) {
	d := r.Daily
	if len(d.Time) != 1 || len(d.WeatherCode) != 1 || len(d.Temperature2mMax) != 1 ||
		len(d.Temperature2mMin) != 1 || len(d.PrecipitationProbabilityMax) != 1 {
		return domain.Forecast{}, fmt.Errorf("daily values for exactly one day are required: got time=%d weather_code=%d temperature_2m_max=%d temperature_2m_min=%d precipitation_probability_max=%d",
			len(d.Time), len(d.WeatherCode), len(d.Temperature2mMax), len(d.Temperature2mMin), len(d.PrecipitationProbabilityMax))
	}
	if d.Time[0] != date.String() {
		return domain.Forecast{}, fmt.Errorf("daily time is %s, want %s", d.Time[0], date)
	}
	if d.WeatherCode[0] == nil || d.Temperature2mMax[0] == nil || d.Temperature2mMin[0] == nil || d.PrecipitationProbabilityMax[0] == nil {
		return domain.Forecast{}, fmt.Errorf("daily values are missing for %s", date)
	}

	weather, err := weatherFromCode(*d.WeatherCode[0])
	if err != nil {
		return domain.Forecast{}, err
	}
	maxTemp, err := domain.NewTemperature(*d.Temperature2mMax[0])
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("temperature_2m_max: %w", err)
	}
	minTemp, err := domain.NewTemperature(*d.Temperature2mMin[0])
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("temperature_2m_min: %w", err)
	}
	precip, err := domain.NewPrecipitationProbability(*d.PrecipitationProbabilityMax[0])
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("precipitation_probability_max: %w", err)
	}
	return domain.NewForecast(location, date, weather, maxTemp, minTemp, precip)
}

// weatherNames は WMO の天気コードから天気区分への対応表。
// Open-Meteo のドキュメントに載っているコードだけを持つ。
var weatherNames = map[int]string{
	0: "clear",
	1: "cloudy", 2: "cloudy", 3: "cloudy",
	45: "fog", 48: "fog",
	51: "rain", 53: "rain", 55: "rain", 56: "rain", 57: "rain",
	61: "rain", 63: "rain", 65: "rain", 66: "rain", 67: "rain",
	80: "rain", 81: "rain", 82: "rain",
	71: "snow", 73: "snow", 75: "snow", 77: "snow",
	85: "snow", 86: "snow",
	95: "thunderstorm", 96: "thunderstorm", 99: "thunderstorm",
}

func weatherFromCode(code int) (domain.Weather, error) {
	name, ok := weatherNames[code]
	if !ok {
		return domain.Weather{}, fmt.Errorf("unsupported weather code %d", code)
	}
	return domain.NewWeather(name)
}
