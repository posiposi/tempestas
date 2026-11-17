package domain

import "errors"

// ErrInvalidWeather は定義されていない天気区分を受け取ったときに返すエラー。
var ErrInvalidWeather = errors.New("invalid weather")

// Weather は予報の天気区分。api/openapi.yaml の Weather の enum に対応する6種類。
// 生成は NewWeather で行う。ゼロ値は有効な天気区分を表さない。
type Weather struct {
	name string
}

var weathers = map[string]Weather{
	"clear":        {name: "clear"},
	"cloudy":       {name: "cloudy"},
	"fog":          {name: "fog"},
	"rain":         {name: "rain"},
	"snow":         {name: "snow"},
	"thunderstorm": {name: "thunderstorm"},
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
