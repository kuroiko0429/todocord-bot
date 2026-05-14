BINARY_NAME=bot
BUILD_DIR=./build

.PHONY: all build run test clean

## build: バイナリをビルドします
build:
	@echo "==> Buildingリ $(BINARY_NAME)..."
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

## help: 利用可能なコマンド一覧を表示します
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
