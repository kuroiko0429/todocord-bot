# ── Stage 1: ビルド ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 依存関係のキャッシュ（ソースより先にコピーしてレイヤーキャッシュを活用）
COPY go.mod go.sum ./
RUN go mod download

# ソースをコピーしてビルド（CGO不要 = modernc.org/sqlite 使用）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bot ./cmd/bot


# ── Stage 2: 実行環境（最小限） ──────────────────────────────────
FROM alpine:latest

# タイムゾーンデータ（期限通知・スケジューラのJST対応に必要）
RUN apk add --no-cache tzdata ca-certificates

# 実行ユーザーを作成（rootで動かさないためのセキュリティ対策）
RUN addgroup -S botgroup && adduser -S botuser -G botgroup

# DBファイルを永続化するディレクトリ
RUN mkdir -p /data && chown botuser:botgroup /data

WORKDIR /app
COPY --from=builder /bot ./bot

USER botuser

# SQLiteファイルは /data 以下に保存（Volumeでマウントすることを想定）
ENV DB_PATH=/data/todocord.db
ENV TZ=Asia/Tokyo

ENTRYPOINT ["./bot"]
