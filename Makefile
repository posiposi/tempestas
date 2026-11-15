BUILD_DIR := build
BINARY    := $(BUILD_DIR)/bootstrap
ZIP       := $(BUILD_DIR)/api.zip

# ビルドは Go 1.27.3 の開発用コンテナで行う（docker-compose.yml の app サービス）。
GO_RUN := docker compose run --rm -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 app

.PHONY: build clean

# Lambda（provided.al2023 / arm64）用の bootstrap を作り、zip にまとめる。
# - lambda.norpc: provided.al2023 では使わない RPC モードのコードを除く
# - -trimpath, -buildvcs=false: ビルド環境の違いでバイナリが変わらないようにする
# - mtime を固定し zip -X で余分な属性を省く: 中身が同じなら zip のハッシュも同じになり、
#   Terraform が変更なしの Lambda を更新しない
# golang イメージには zip が入っていないため、zip はホストで実行する。
# mtime の変更はファイルの所有者でないとできない（コンテナは root で動く）ので、コンテナ内で行う。
# Linux ホストではコンテナが作ったファイルは root 所有になる。build/ をホストで先に作っておくことで、
# ホスト側から zip の書き込みや make clean での削除ができるようにする。
build:
	mkdir -p $(BUILD_DIR)
	$(GO_RUN) sh -c 'go build -tags lambda.norpc -trimpath -buildvcs=false -ldflags="-s -w" -o $(BINARY) ./cmd/api && touch -d 1980-01-01T00:00:00Z $(BINARY)'
	rm -f $(ZIP)
	TZ=UTC zip -X -j $(ZIP) $(BINARY)

clean:
	rm -rf $(BUILD_DIR)
