// Command api は tempestas の Lambda エントリポイント。
// API Gateway（HTTP API, ペイロード形式 2.0）からのイベントを、
// aws-lambda-go-api-proxy で http.Handler へのリクエストに変換する。
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/posiposi/tempestas/internal/adapter/handler"
)

func main() {
	adapter := httpadapter.NewV2(handler.New())
	lambda.Start(adapter.ProxyWithContext)
}
