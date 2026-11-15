BUILD_DIR := build
BINARY    := $(BUILD_DIR)/bootstrap
ZIP       := $(BUILD_DIR)/api.zip

# ビルドは開発用コンテナで行う（docker-compose.yml の app サービス、docker/app/Dockerfile）。
# --build: Dockerfile を変えたときに、古いイメージのままビルドしないようにする。
# GOTOOLCHAIN=local: go.mod がコンテナより新しい Go を求めたとき、別のツールチェーンを
# 自動で取得せずに失敗させる（ビルドに使う Go をコンテナのバージョンに固定するため）。
GO_RUN := docker compose run --rm --build -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 -e GOTOOLCHAIN=local app

.PHONY: build clean

# Lambda（provided.al2023 / arm64）用の bootstrap を作り、zip にまとめる。
# - lambda.norpc: provided.al2023 では使わない RPC モードのコードを除く
# - -trimpath, -buildvcs=false: ビルド環境の違いでバイナリが変わらないようにする
# - mtime を固定し zip -X で余分な属性を省く: 中身が同じなら zip のハッシュも同じになり、
#   Terraform が変更なしの Lambda を更新しない
# Linux ホストではコンテナが作ったファイルは root 所有になる。build/ をホストで先に作っておくことで、
# make clean で削除できるようにする。
build:
	mkdir -p $(BUILD_DIR)
	$(GO_RUN) sh -c '\
		go build -tags lambda.norpc -trimpath -buildvcs=false -ldflags="-s -w" -o $(BINARY) ./cmd/api && \
		touch -d 1980-01-01T00:00:00Z $(BINARY) && \
		rm -f $(ZIP) && \
		zip -X -j $(ZIP) $(BINARY)'

clean:
	rm -rf $(BUILD_DIR)
