// Package handler は tempestas の HTTP ハンドラを提供する。
// Lambda（cmd/api）とローカルサーバー（cmd/local）の両方から同じ http.Handler を使うため、
// Lambda 固有の処理はここに入れない。
package handler

import (
	"net/http"
)

// helloWeatherJSON はデプロイ経路を確認するための固定レスポンス（api/openapi.yaml の Forecast の例と同じ内容）。
// 後段のStepで固定レスポンスは削除する予定
const helloWeatherJSON = `{"location":"tokyo","date":"2026-11-07","weather":"clear","temperature":{"max":18.2,"min":9.5},"precipitationProbability":10}`

// New は API のルーティングを設定した http.Handler を返す。
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /weather", getWeather)
	return mux
}

func getWeather(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(helloWeatherJSON))
}
