# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

tempestas は、地域・日付を指定して天気予報を返すサーバーサイド専用のサービス（Go + AWS Lambda）。フロントエンドはなく、インフラはすべて Terraform で管理する。転職活動用のポートフォリオなので、**設計判断を説明しやすいこと**を重視する。機能を足すときは、なぜその方法を選んだのかをコードの構成やREADMEで説明できるようにする。

## 現在のフェーズ

- **MVPは `GET /weather` のみ。** 購読API、EventBridge Scheduler、SQS、通知Lambda、Slack連携は次のフェーズで扱うため、頼まれない限り実装しない。
- 進め方はウォーキングスケルトン方式（下記「進め方」の Step 1〜7）。まず Hello World を本番環境までデプロイし、IAM・ビルド・デプロイの経路を確立してから中身を足す。
- 現時点では `cmd/api`（固定のJSONを返す Hello World の Lambda）、`internal/adapter/handler`、`make build`、`terraform/envs/dev`（API Gateway + Lambda の定義。スロットリングは rate 5 / burst 10）のみ実装済みで、Hello World は dev 環境にデプロイして疎通を確認済み。下記のコマンドや構成のうち未実装のものは計画段階のもので、実装したら実際の内容に合わせてこのファイルを更新する。

## 進め方

本番環境まで通る最小構成を先に作り（ウォーキングスケルトン）、そこに中身を足していく。

1. **API仕様を決める**
   - `api/openapi.yaml` を作成する
   - 対応する地域コードと座標の一覧を決める
2. **Terraformの土台と Hello World のデプロイ**
   - `terraform/bootstrap` で state 用の S3 バケットを作成する
   - API Gateway（HTTP API）+ Lambda（固定のJSONを返すだけ）を構築する
   - `make build`（`GOOS=linux GOARCH=arm64`、`bootstrap` バイナリを生成）を整備する
   - curl で疎通を確認し、IAM・ビルド・デプロイの経路を確立する
3. **ドメインとユースケース（TDD）**
   - `Location`、`ForecastDate`（予報範囲のバリデーションを含む）を実装する
   - `GetWeather` ユースケース：キャッシュがあれば返し、なければ取得して保存する
   - `WeatherFetcher` / `ForecastRepository` をインターフェースとして定義し、モックでテストする
4. **アダプタの実装**
   - Open-Meteo クライアント（タイムアウトを設定し、`httptest` でテスト）
   - DynamoDB リポジトリ（testcontainers で結合テスト、TTL 属性を設定）
   - HTTP ハンドラ（バリデーションエラーは 400、外部APIのエラーは 502）
5. **組み立てと再デプロイ**
   - Terraform に DynamoDB テーブル（TTL 有効）と最小権限の IAM ポリシーを追加する
   - Hello World を本物のハンドラに差し替える
   - CloudWatch Logs で構造化ログを確認する
6. **CI/CD の整備**
   - PR 時：golangci-lint、テスト、`terraform fmt -check` / `validate` / `plan`
   - main へのマージ時：ビルドと `terraform apply`
   - AWS の認証は OIDC（IAM ロールを AssumeRole する）
7. **README の仕上げ**
   - 構成図と API 仕様へのリンク
   - 技術選定の理由（DynamoDB や Lambda を選んだ理由など）
   - ローカルでの起動手順（`docker compose up` → `make run`）
   - 今後の拡張予定（通知フロー）

## コマンド（予定）

```sh
docker compose up -d           # Swagger UI（http://localhost:8081、api/openapi.yaml を表示）と DynamoDB Local（予定）を起動
make run                       # cmd/local のHTTPサーバーをローカルで起動
make build                     # 開発用コンテナで GOOS=linux GOARCH=arm64 の bootstrap をビルドし、build/api.zip にまとめる（Lambda provided.al2023 用）
make clean                     # build/ を削除
go test ./...                  # 全テスト
go test ./internal/domain -run TestForecastDate   # 単一テストを実行
golangci-lint run              # 静的解析
terraform -chdir=terraform/envs/dev fmt -check && terraform -chdir=terraform/envs/dev validate && terraform -chdir=terraform/envs/dev plan
```

### デプロイ（dev 環境）

`plan` / `apply` は `build/api.zip` のハッシュを読むので、先に `make build` を実行する。AWS への認証は `aws sso login` で行う（`default` プロファイルを使うので `AWS_PROFILE` の指定は不要）。

```sh
make build
terraform -chdir=terraform/envs/dev plan
terraform -chdir=terraform/envs/dev apply
terraform -chdir=terraform/envs/dev output weather_url     # API の URL を表示
curl -i <weather_url>                                       # 200 が返ることを確認
aws logs tail /aws/lambda/tempestas-dev-api --since 30m    # Lambda のログを確認
```

DynamoDB の結合テストは testcontainers（DynamoDB Local）を使うため、Docker が起動している必要がある。

### ツールのバージョン

- **Go 1.27.3:** ローカルの Go は gopls などエディタ用にだけ使い、ビルド・テスト・lint は Docker のコンテナ（Go 1.27.3）の中で行う。上記の `go` / `golangci-lint` のコマンドは、コンテナの中で実行する想定。`make` はホストで実行し、ビルドなどは Makefile の中からコンテナを呼び出す。開発用コンテナは `docker-compose.yml` の `app` サービス（profile `tools` のため `docker compose up` では起動しない）で、`docker compose run --rm app go build ./...` のように使う。イメージは `docker/app/Dockerfile`（`golang:1.27.3` に `zip` を追加したもの）。ソースコードはバインドマウントでホストと共有しているので、編集はホストで行う。
- **Terraform 1.17.1:** ローカルにインストールしたものを使う。
- **make:** ホストに必要（macOS と GitHub Actions の Ubuntu ランナーには最初から入っている）。

## アーキテクチャ

クリーンアーキテクチャ風のレイヤー構成。依存の向きは `adapter → usecase → domain`。

- `internal/domain` — 値オブジェクト `Location`（固定リストにある地域コードのみ有効）と `ForecastDate`（YYYY-MM-DD、JSTの当日〜6日後のみ有効）、エンティティ `Forecast`。入力のバリデーションはここに置く。緯度経度ではなく地域コードで受け付けるのは、地域を値オブジェクトとしてドメインに閉じ込めるため。
- `internal/usecase` — `GetWeather` ユースケース。ポートとして `WeatherFetcher`（外部API）と `ForecastRepository`（キャッシュ）のインターフェースを定義する。処理の流れは、キャッシュにあれば返し、なければ外部APIから取得して保存してから返す。ユニットテストではポートをモックにする（TDDで進める）。
- `internal/adapter/handler` — 標準 `net/http` のハンドラ。HTTPフレームワークは使わない。エラーの対応は、ドメインのバリデーションエラーを 400（`INVALID_LOCATION` / `INVALID_DATE`）、外部APIの失敗（気温・降水確率など必要な値が `null` の場合を含む）を 502（`BAD_GATEWAY`）、キャッシュ（DynamoDB）の読み書きの失敗などそれ以外のエラーを 500（`INTERNAL_SERVER_ERROR`）とする。
- `internal/adapter/openmeteo` — Open-Meteo（APIキー不要）のクライアント。タイムアウトを必ず設定する。テストには `httptest` を使う。
- `internal/adapter/dynamodb` — `aws-sdk-go-v2` で実装するリポジトリ。
- `cmd/api`（Lambdaエントリポイント）と `cmd/local`（ローカル用HTTPサーバー）は、**同じ `http.Handler`** を使う。Lambda では `aws-lambda-go-api-proxy` で包む。Lambda 専用のロジックをハンドラに入れないこと。
- `api/openapi.yaml` — APIの仕様はここを正とし、仕様を先に決めてから実装する。エンドポイントやレスポンスを変えるときは、OpenAPIを先に更新する。

### DynamoDB

- テーブル `weather-notify`。次のフェーズで購読情報も同じテーブルに入れる**シングルテーブル設計**を前提にしている。
- キーは `PK = LOC#<location>`、`SK = DATE#<YYYY-MM-DD>`。MVPのアクセスパターンは `GetItem` 1種類だけ。
- `expiresAt`（UNIX秒）は TTL 属性で、取得からおよそ3時間後に設定する。

### インフラ / CI

- リージョンは東京（`ap-northeast-1`）。ローカルから AWS への認証は IAM Identity Center（SSO）のプロファイルを使い、IAM ユーザーのアクセスキーは使わない。
- ほかのリソースと区別するため、Terraform で作るリソースにはすべて `Project = tempestas` のタグを付ける（provider の `default_tags` で設定する）。
- Lambda は `provided.al2023` / arm64、API Gateway は HTTP API（REST APIではない）、DynamoDB はオンデマンド課金。
- Terraform の state は S3 backend に置き、`use_lockfile` でロックする（DynamoDBのロックテーブルは使わない）。`terraform/bootstrap` で state 用のバケットを作り、`terraform/envs/dev` で環境を構築する。モジュールへの分割は通知フローを追加するときまで行わない。
- GitHub Actions から AWS への認証は OIDC で行い、長期のアクセスキーは使わない。PRでは lint・テスト・`terraform fmt -check`/`validate`/`plan` を実行し、main へのマージ時にビルドと `apply` を実行する。
- IAM ポリシーは最小権限で書く。

### ログ

`log/slog` を使い、JSON形式で出力する（CloudWatch Logs Insights で検索できるようにするため）。Lambda の `log_format` も `JSON` にしているので、Lambda が出す `platform.start` / `platform.report` などのログも JSON になり、アプリのログと同じ形式で検索できる。

### コメント（godoc）

[Go Doc Comments](https://go.dev/doc/comment) と [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#doc-comments) の慣習に従う。意味はまず名前と型で表し、コメントは名前と型で表せないことの補足に限る。

- パッケージコメントは必ず書く（`// Package handler は…`）。`main` パッケージは `// Command api は…` の形にする。
- エクスポートした識別子には doc コメントを書き、識別子名から書き始める（`// New は…`）。
- 非公開の識別子には原則書かない。込み入ったものや、理由がコードから読み取れないもの（一時的な実装、いつ消すかなど）にだけ書く。
- `@param` / `@return` のようなタグは使わず、文章で書く。引数や戻り値は名前でそのまま参照し、戻り値が複数で説明しにくいときは戻り値に名前を付ける。
- 書くのは、エラーを返す条件、値の範囲や単位、`nil` やゼロ値の意味、副作用、並行に呼んでも安全かなど、使う側が知っておくべき契約。名前を言い換えるだけのコメント（`// GetWeather は天気を取得する。`）は書かない。
