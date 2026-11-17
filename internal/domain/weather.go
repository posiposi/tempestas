package domain

import "errors"

// ErrInvalidWeather は定義されていない天気区分を受け取ったときに返すエラー。
var ErrInvalidWeather = errors.New("invalid weather")

// Weather は予報の天気区分。
// 値は WeatherClear などの定義済みの変数か NewWeather で得る。ゼロ値は有効な天気区分を表さない。
// 比較可能なので == や switch で比べられる。
type Weather struct {
	name string
}

// 天気区分の定義済みの値。api/openapi.yaml の Weather の enum に対応する。
// パッケージの外から再代入しないこと。
var (
	WeatherClear        = Weather{name: "clear"}
	WeatherCloudy       = Weather{name: "cloudy"}
	WeatherFog          = Weather{name: "fog"}
	WeatherRain         = Weather{name: "rain"}
	WeatherSnow         = Weather{name: "snow"}
	WeatherThunderstorm = Weather{name: "thunderstorm"}
)

var weathers = map[string]Weather{
	WeatherClear.name:        WeatherClear,
	WeatherCloudy.name:       WeatherCloudy,
	WeatherFog.name:          WeatherFog,
	WeatherRain.name:         WeatherRain,
	WeatherSnow.name:         WeatherSnow,
	WeatherThunderstorm.name: WeatherThunderstorm,
}

// NewWeather は文字列 s が表す Weather を返す。
// s は String が返す小文字の値（例: "clear"）で、大文字や前後の空白は受け付けない。
// 定義されていない値の場合は ErrInvalidWeather を返す。
func NewWeather(s string) (Weather, error) {
	w, ok := weathers[s]
	if !ok {
		return Weather{}, ErrInvalidWeather
	}
	return w, nil
}

// String は天気区分を表す小文字の文字列（例: "clear"）を返す。ゼロ値では空文字を返す。
func (w Weather) String() string { return w.name }
