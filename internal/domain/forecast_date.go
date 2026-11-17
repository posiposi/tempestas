package domain

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidDate は予報日の形式が不正か、予報できる範囲外の日付を受け取ったときに返すエラー。
var ErrInvalidDate = errors.New("invalid date")

// forecastDays は予報できる日数で、当日を含めて7日（当日〜6日後）。
const forecastDays = 7

const dateLayout = "2006-01-02"

// jst は日本標準時。日本は1951年以降夏時間がなく UTC+9 で固定のため、
// tzdata に依存しない固定オフセットで表す。
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// ForecastDate は予報の対象日（日本時間の暦日）。
// 有効な値は NewForecastDate か TodayForecastDate で生成する。ゼロ値は有効な予報日を表さない。
// 年月日だけを持つため == で比較できる。
type ForecastDate struct {
	year  int
	month time.Month
	day   int
}

// NewForecastDate は YYYY-MM-DD 形式の s から ForecastDate を返す。
// 有効な範囲は now を日本時間に変換した日の当日〜6日後で、now のタイムゾーンは問わない。
// 月日はゼロ埋めの2桁が必要で、前後の空白や存在しない日付は受け付けない。
// 形式が不正な場合と範囲外の場合は ErrInvalidDate を返す。
func NewForecastDate(s string, now time.Time) (ForecastDate, error) {
	t, err := time.ParseInLocation(dateLayout, s, jst)
	if err != nil {
		return ForecastDate{}, fmt.Errorf("%w: %q is not YYYY-MM-DD", ErrInvalidDate, s)
	}
	today := TodayForecastDate(now).midnight()
	last := today.AddDate(0, 0, forecastDays-1)
	if t.Before(today) || t.After(last) {
		return ForecastDate{}, fmt.Errorf("%w: %s is out of range %s to %s", ErrInvalidDate, s, today.Format(dateLayout), last.Format(dateLayout))
	}
	return fromTime(t), nil
}

// TodayForecastDate は now を日本時間に変換した日の ForecastDate を返す。
// date を省略したリクエストの既定値に使う。
func TodayForecastDate(now time.Time) ForecastDate {
	return fromTime(now.In(jst))
}

// String は予報日を YYYY-MM-DD 形式で返す。
func (d ForecastDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func fromTime(t time.Time) ForecastDate {
	y, m, d := t.Date()
	return ForecastDate{year: y, month: m, day: d}
}

func (d ForecastDate) midnight() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, jst)
}
