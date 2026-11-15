# 開発用コンテナ（docker-compose.yml の app サービス）のイメージ。
# zip: make build で Lambda 用の zip を作る（ホストの zip に頼らず、どこで作っても同じ zip にするため）
FROM golang:1.27.3

RUN apt-get update \
    && apt-get install -y --no-install-recommends zip \
    && rm -rf /var/lib/apt/lists/*
