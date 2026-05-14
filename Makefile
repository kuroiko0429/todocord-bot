BINARY_NAME=bot
BUILD_DIR=./build
IMAGE_NAME=todocord-bot

.PHONY: all build run test clean setup \
        docker-build docker-up docker-down docker-logs docker-restart docker-clean

# ── ローカル開発 ───────────────────────────────────────────────────

## build: バイナリをビルドします
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/bot

## run: 開発用にそのままGoで起動します (.envが必要です)
run:
	go run ./cmd/bot

## test: ユニットテストを実行します
test:
	go test -v -count=1 ./...

## clean: ビルド成果物を削除します
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BUILD_DIR)

## setup: .envファイルをサンプルから作成します（初回のみ）
setup:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ .env ファイルを作成しました。DISCORD_TOKEN等を設定してください。"; \
	else \
		echo "⚠️  .env はすでに存在します。スキップしました。"; \
	fi

# ── Docker 操作 ────────────────────────────────────────────────────

## docker-build: Dockerイメージをビルドします
docker-build:
	docker compose build

## docker-up: コンテナをバックグラウンドで起動します
docker-up:
	docker compose up -d
	@echo "✅ Botが起動しました。ログは 'make docker-logs' で確認できます。"

## docker-down: コンテナを停止・削除します（データは保持）
docker-down:
	docker compose down

## docker-logs: コンテナのログをリアルタイムで表示します
docker-logs:
	docker compose logs -f

## docker-restart: コンテナを再起動します
docker-restart:
	docker compose restart
	@echo "✅ Botを再起動しました。"

## docker-rebuild: イメージを再ビルドしてコンテナを起動します（コード変更後）
docker-rebuild:
	docker compose down
	docker compose build --no-cache
	docker compose up -d
	@echo "✅ Botを再ビルドして起動しました。"

## docker-clean: コンテナ・イメージ・ボリュームを全て削除します（データも消えます！）
docker-clean:
	@echo "⚠️  データボリュームを含む全てのリソースを削除します..."
	docker compose down -v --rmi all

# ── ヘルプ ────────────────────────────────────────────────────────

## help: 利用可能なコマンド一覧を表示します
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
