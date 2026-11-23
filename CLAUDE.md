# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

tempestas は、地域・日付を指定して天気予報を返すサーバーサイド専用のサービス（Go + AWS Lambda）。フロントエンドはなく、インフラはすべて Terraform で管理する。転職活動用のポートフォリオなので、**設計判断を説明しやすいこと**を重視する。機能を足すときは、なぜその方法を選んだのかをコードの構成やREADMEで説明できるようにする。

## 現在のフェーズ

- **MVPは `GET /weather` のみ。** 購読API、EventBridge（Scheduler とカスタムバス）、SQS、取得Lambda、通知Lambda、SES によるメール通知は次のフェーズで扱うため、頼まれない限り実装しない。完成形の通知フローは下記「通知フロー（次のフェーズ）」のとおり。
- 進め方はウォーキングスケルトン方式（下記「進め方」の Step 1〜7）。まず Hello World を本番環境までデプロイし、IAM・ビルド・デプロイの経路を確立してから中身を足す。
- 現時点では `cmd/api`（固定のJSONを返す Hello World の Lambda）、`internal/adapter/handler`、`make build`、`terraform/envs/dev`（API Gateway + Lambda の定義。スロットリングは rate 5 / burst 10）、`internal/domain`（値オブジェクトとエンティティ `Forecast`）のみ実装済みで、Hello World は dev 環境にデプロイして疎通を確認済み。下記のコマンドや構成のうち未実装のものは計画段階のもので、実装したら実際の内容に合わせてこのファイルを更新する。

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
   - 値オブジェクト `Location`、`ForecastDate`（予報範囲のバリデーションを含む）、`Weather`、`Temperature`、`PrecipitationProbability` と、エンティティ `Forecast` を実装する（実装済み）
   - `GetDailyForecast` ユースケース：キャッシュがあれば返し、なければ取得して保存する
   - `ForecastCache` / `WeatherRepository` をインターフェースとして定義し、モックでテストする
4. **アダプタの実装**
   - Open-Meteo クライアント（タイムアウトを設定し、`httptest` でテスト）
   - DynamoDB のキャッシュ（testcontainers で結合テスト、TTL 属性を設定）
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

クリーンアーキテクチャ風のレイヤー構成。依存の向きは「アダプタ → ユースケース → ドメイン」で、内側のレイヤーは外側を知らない。

- **ドメイン** — 値オブジェクトとエンティティ。入力のバリデーションはここに置く。
  - 値1つの制約は値オブジェクト、値どうしの関係はエンティティで守る。
  - ドメインモデルはファクトリ関数でだけ生成する。フィールドは非公開にし、ゼロ値を直接使う前提のコードやテストは書かない。
  - エラーはセンチネルエラーで返し、外側のレイヤーが `errors.Is` で判別する。利用者の入力を検証するエラーと、外部APIから受け取った値を検証するエラーは区別できるようにする。
  - 現在時刻は引数で受け取り、ドメインの中で取得しない。タイムゾーンは tzdata に依存しない固定オフセットで表す。
  - 外部APIの表現（コード値など）からドメインの値への変換は外部APIの都合なので、ドメインではなくアダプタで行う。
  - 外部からの入力は座標などの生の値ではなく、ドメインで定めた識別子（地域コードなど）で受け付け、値オブジェクトに閉じ込める。
- **ユースケース** — アプリケーションの処理の流れ。外部APIとキャッシュはポート（インターフェース）として、使う側であるユースケースに定義し、アダプタが実装する。
  - 外部APIの呼び出しには料金がかかるため、呼び出し回数を抑えることを要件とし、キャッシュを使うかの判断はユースケースで行う。キャッシュにあれば外部APIを呼ばずに返し、なければ外部APIから取得してキャッシュに保存してから返す。
  - キャッシュの読み込みに失敗したときは、外部APIの呼び出しが増えるのを避けるため、外部APIは呼ばずにエラーを返す。
  - キャッシュの保存先やその技術（期限の設定方法など）はアダプタの都合なので、ユースケースは知らない。
  - APIの入口からも次のフェーズのバッチ（定期実行）からも同じユースケースを呼ぶため、HTTP や Lambda などの入口に依存する処理を入れない。通知のためのイベントの発行もユースケースには入れない（理由は「通知フロー（次のフェーズ）」）。
  - ユニットテストではポートを偽物に差し替える（TDDで進める）。
- **アダプタ** — 入口（HTTP ハンドラ）と、ポートの実装（外部APIのクライアント、キャッシュ）。
  - HTTP ハンドラは標準ライブラリで書き、HTTPフレームワークは使わない。ユースケースは、ハンドラ側に定義した非公開のインターフェース越しに呼ぶ。
  - エラーと HTTP ステータスの対応は、利用者の入力のバリデーションエラーを 400、外部APIの失敗（必要な値が欠けている場合や、受け取った値がドメインの検証を通らない場合を含む）を 502、キャッシュの読み書きの失敗などそれ以外を 500 とする。
  - 外部APIのクライアントには必ずタイムアウトを設定し、テストではスタブのHTTPサーバーを使う。
- **エントリポイント** — Lambda とローカル用のサーバーは同じ HTTP ハンドラを使う。Lambda のイベントの変換はエントリポイントに閉じ込め、Lambda 専用のロジックをハンドラに入れない。
- **API仕様** — OpenAPI の定義を正とし、仕様を先に決めてから実装する。エンドポイントやレスポンスを変えるときは、OpenAPI を先に更新する。

### DynamoDB

- テーブル `weather-notify`。次のフェーズで購読情報も同じテーブルに入れる**シングルテーブル設計**を前提にしている。
- キーは `PK = LOC#<location>`、`SK = DATE#<YYYY-MM-DD>`。MVPのアクセスパターンは `GetItem` 1種類だけ。
- `expiresAt`（UNIX秒）は TTL 属性で、取得からおよそ3時間後に設定する。
- 次のフェーズでは、通知Lambdaが地域から購読者を引くアクセスパターンと、通知済みを記録する冪等キー（購読ID + 日付）が加わる。購読情報のキー設計は購読APIを設計するときに決める。

### 通知フロー（次のフェーズ）

完成形は**イベント駆動**で、その日の予報が用意できたタイミングで購読者に通知する。構成図は `docs/full-architecture.drawio`（README には書き出した `docs/full-architecture.svg` を載せる）。

1. EventBridge Scheduler が毎朝、取得Lambdaを起動する。
2. 取得Lambdaは対応しているすべての地域について、その日の予報を地域ごとに1回だけ `GetDailyForecast` で用意し、地域ごとに `ForecastReady`（`location`、`date` と予報の中身）を EventBridge のカスタムバスに発行する。
3. EventBridge のルールがイベントを SQS（DLQ 付き）に送る。
4. 通知Lambdaはイベントの地域の購読者を DynamoDB から `Query` し、購読者1人ずつ SES（`SendEmail`）でメールを送る。

- 取得Lambdaは購読者も通知先も知らない。誰にどう通知するかはイベントを受け取る側が決める。受け取る処理を増やすときはルールを足し、取得Lambdaは変えない。
- 取得Lambdaは購読情報を読まない。購読されている地域だけに絞らず、固定リストのすべての地域のイベントを発行する。購読者がいない地域は通知Lambdaが何もせずに終える。
- イベントに予報の中身を入れるので、通知Lambdaは予報のキャッシュを読まない（TTL で消えたあとに再処理されても困らないようにするため）。
- `ForecastReady` の意味は「その日の予報が用意できた」であり、「外部APIから取得した」ではない。`GetDailyForecast` はキャッシュを先に見るので、事前に `GET /weather` が呼ばれていると外部APIへの取得が起きない。そのためイベントはユースケースの中ではなく、取得Lambdaが予報を手に入れたあとに、キャッシュから取ったかどうかに関係なく発行する。
- 通知Lambdaは1つのイベントで地域の購読者全員に送るので、再配信（EventBridge と SQS は少なくとも1回の配信）に備えて購読ID + 日付の冪等キーで二重送信を防ぐ。購読者ごとに「冪等キーがあればスキップ → SES で送信 → 冪等キーを記録（TTL 付き）」の順で処理する。記録を送信のあとにするのは、送信に失敗した購読者に届かなくなるのを避けるため（二重に届くのは送信と記録の間で止まったときだけ）。
- 宛先は1通に全員を入れず、購読者1人ずつ送る（アドレスを互いに見せないためと、冪等キーを購読者ごとに記録するため）。
- Slack への通知は行わない。

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
