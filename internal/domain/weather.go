package domain

import "errors"

// ErrInvalidWeather は定義されていない天気区分を受け取ったときに返すエラー。
var ErrInvalidWeather = errors.New("invalid weather")

// Weather は予報の天気区分。
// 値は WeatherClear などの関数か NewWeather で得る。ゼロ値は有効な天気区分を表さない。
// 定義済みの値は api/openapi.yaml の Weather の enum に対応する6種類。
type Weather struct {
	name string
}

// WeatherClear は晴れ（"clear"）を返す。
func WeatherClear() Weather { return Weather{name: "clear"} }

// WeatherCloudy は曇り（"cloudy"）を返す。
func WeatherCloudy() Weather { return Weather{name: "cloudy"} }

// WeatherFog は霧（"fog"）を返す。
func WeatherFog() Weather { return Weather{name: "fog"} }

// WeatherRain は雨（"rain"）を返す。
func WeatherRain() Weather { return Weather{name: "rain"} }

// WeatherSnow は雪（"snow"）を返す。
func WeatherSnow() Weather { return Weather{name: "snow"} }

// WeatherThunderstorm は雷雨（"thunderstorm"）を返す。
func WeatherThunderstorm() Weather { return Weather{name: "thunderstorm"} }

var weathers = map[string]Weather{
	WeatherClear().name:        WeatherClear(),
	WeatherCloudy().name:       WeatherCloudy(),
	WeatherFog().name:          WeatherFog(),
	WeatherRain().name:         WeatherRain(),
	WeatherSnow().name:         WeatherSnow(),
	WeatherThunderstorm().name: WeatherThunderstorm(),
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
