package domain

import (
	"errors"
	"math"
)

// ErrInvalidTemperature は NaN または ±Inf の気温を受け取ったときに返すエラー。
var ErrInvalidTemperature = errors.New("invalid temperature")

// Temperature は気温（℃）。
// 生成は NewTemperature で行う。0℃ は有効な気温のため、ゼロ値は 0℃ を表す。
// 値は常に有限なので、2つの Temperature は Celsius の値を比べて大小を判定できる。
type Temperature struct {
	celsius float64
}

// NewTemperature は celsius（℃）の Temperature を返す。
// 気温には根拠のある上限・下限がないため範囲は制限せず、
// celsius が NaN または ±Inf の場合にだけ ErrInvalidTemperature を返す。
func NewTemperature(celsius float64) (Temperature, error) {
	if math.IsNaN(celsius) || math.IsInf(celsius, 0) {
		return Temperature{}, ErrInvalidTemperature
	}
	return Temperature{celsius: celsius}, nil
}

// Celsius は気温を℃で返す。値は常に有限。
func (t Temperature) Celsius() float64 { return t.celsius }
