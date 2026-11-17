package domain

import "errors"

// ErrPrecipitationProbabilityOutOfRange は降水確率が 0〜100 の範囲外のときに返すエラー。
// 降水確率は利用者の入力ではなく外部 API から受け取る値なので、このエラーは外部 API の異常を表す。
// 利用者の入力ミス（400）としては扱わず、外部 API の失敗（502）として扱う。
var ErrPrecipitationProbabilityOutOfRange = errors.New("precipitation probability out of range")

// PrecipitationProbability は降水確率（%）。
// 生成は NewPrecipitationProbability で行う。
// ゼロ値は 0% を表す有効な値。
type PrecipitationProbability struct {
	percent int
}

const (
	minPrecipitationProbability = 0
	maxPrecipitationProbability = 100
)

// NewPrecipitationProbability は percent（%）を降水確率として返す。
// percent が 0〜100 の範囲外の場合は ErrPrecipitationProbabilityOutOfRange を返す。
func NewPrecipitationProbability(percent int) (PrecipitationProbability, error) {
	if percent < minPrecipitationProbability || percent > maxPrecipitationProbability {
		return PrecipitationProbability{}, ErrPrecipitationProbabilityOutOfRange
	}
	return PrecipitationProbability{percent: percent}, nil
}

// Percent は降水確率を 0〜100 の整数（%）で返す。
func (p PrecipitationProbability) Percent() int { return p.percent }
