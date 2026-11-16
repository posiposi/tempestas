// Package domain は天気予報のドメインモデル（値オブジェクトとエンティティ）を提供する。
package domain

import "errors"

// ErrInvalidLocation は対応していない地域コードを受け取ったときに返すエラー。
var ErrInvalidLocation = errors.New("invalid location")

// Location は予報を提供する地域。
// 生成は NewLocation で行う。ゼロ値は有効な地域を表さない。
// 座標は外部からの入力ではなく、このパッケージで定める予報地点の定義のため float64 で値として持つ。
type Location struct {
	code      string
	latitude  float64
	longitude float64
}

// 予報地点は各市役所（東京は都庁、名古屋は栄交差点）。
var locations = map[string]Location{
	"sapporo": {code: "sapporo", latitude: 43.0621, longitude: 141.3544},
	"tokyo":   {code: "tokyo", latitude: 35.6895, longitude: 139.6917},
	"nagoya":  {code: "nagoya", latitude: 35.1688, longitude: 136.9068},
	"osaka":   {code: "osaka", latitude: 34.6937, longitude: 135.5022},
	"fukuoka": {code: "fukuoka", latitude: 33.5902, longitude: 130.4017},
}

// NewLocation は地域コード code に対応する Location を返す。
// code は小文字のコードで大文字や前後の空白は受け付けない。
// 対応していないコードの場合は ErrInvalidLocation を返す。
func NewLocation(code string) (Location, error) {
	loc, ok := locations[code]
	if !ok {
		return Location{}, ErrInvalidLocation
	}
	return loc, nil
}

// Code は地域コード（例: "tokyo"）を返す。
func (l Location) Code() string { return l.code }

// Latitude は予報地点の緯度（度、北緯が正）を返す。
func (l Location) Latitude() float64 { return l.latitude }

// Longitude は予報地点の経度（度、東経が正）を返す。
func (l Location) Longitude() float64 { return l.longitude }
