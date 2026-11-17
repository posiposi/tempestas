package domain

import "errors"

// ErrInvertedTemperatureRange は最高気温が最低気温より低いときに NewForecast が返すエラー。
// 各気温はそれぞれ有効でも、組み合わせとして矛盾している状態を表す。
// 予報の値は外部 API から取得するので、このエラーは利用者の入力ミスではなく
// 外部 API が矛盾した値を返したことを意味する（handler では 502 に対応付ける）。
var ErrInvertedTemperatureRange = errors.New("max temperature is below min temperature")

// Forecast は1地域・1日分の天気予報を表すエンティティ。
// 生成は NewForecast で行う。ゼロ値は有効な予報を表さない。
type Forecast struct {
	location                 Location
	date                     ForecastDate
	weather                  Weather
	maxTemperature           Temperature
	minTemperature           Temperature
	precipitationProbability PrecipitationProbability
}

// NewForecast は location の date の予報を返す。
// maxTemperature と minTemperature は同じ型なので、引数の順番（最高、最低）に注意する。
// 最高気温と最低気温が等しい場合は有効とし、
// maxTemperature が minTemperature より低い場合は ErrInvertedTemperatureRange を返す。
// 各引数はそれぞれのコンストラクタで生成した値であることを前提とし、ゼロ値かどうかは検証しない。
func NewForecast(
	location Location,
	date ForecastDate,
	weather Weather,
	maxTemperature, minTemperature Temperature,
	precipitationProbability PrecipitationProbability,
) (Forecast, error) {
	if maxTemperature.Celsius() < minTemperature.Celsius() {
		return Forecast{}, ErrInvertedTemperatureRange
	}
	return Forecast{
		location:                 location,
		date:                     date,
		weather:                  weather,
		maxTemperature:           maxTemperature,
		minTemperature:           minTemperature,
		precipitationProbability: precipitationProbability,
	}, nil
}

// Location は予報の地域を返す。
func (f Forecast) Location() Location { return f.location }

// Date は予報日（JST）を返す。
func (f Forecast) Date() ForecastDate { return f.date }

// Weather は予報日の天気を返す。
func (f Forecast) Weather() Weather { return f.weather }

// MaxTemperature は予報日の最高気温を返す。MinTemperature 以上であることが保証される。
func (f Forecast) MaxTemperature() Temperature { return f.maxTemperature }

// MinTemperature は予報日の最低気温を返す。MaxTemperature 以下であることが保証される。
func (f Forecast) MinTemperature() Temperature { return f.minTemperature }

// PrecipitationProbability は予報日の最大降水確率を返す。
func (f Forecast) PrecipitationProbability() PrecipitationProbability {
	return f.precipitationProbability
}
