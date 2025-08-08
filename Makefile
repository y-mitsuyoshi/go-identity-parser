# Go Identity Parser - OCR Web API Makefile

# 変数定義
PROJECT_NAME := go-identity-parser
DOCKER_COMPOSE := docker compose
DOCKER_COMPOSE_DEV := docker compose -f docker-compose.yml
SERVICE_NAME := ocr-api-dev
TEST_SERVICE := test

# デフォルトターゲット
.PHONY: help
help: ## このヘルプを表示
	@echo "Go Identity Parser - OCR Web API"
	@echo ""
	@echo "利用可能なコマンド:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 開発環境
.PHONY: dev
dev: ## 開発環境を起動（ホットリロード付き）
	$(DOCKER_COMPOSE_DEV) up --build $(SERVICE_NAME)

.PHONY: dev-bg
dev-bg: ## 開発環境をバックグラウンドで起動
	$(DOCKER_COMPOSE_DEV) up --build -d $(SERVICE_NAME)

.PHONY: dev-logs
dev-logs: ## 開発環境のログを表示
	$(DOCKER_COMPOSE_DEV) logs -f $(SERVICE_NAME)

.PHONY: dev-stop
dev-stop: ## 開発環境を停止
	$(DOCKER_COMPOSE_DEV) stop

.PHONY: dev-down
dev-down: ## 開発環境を削除（コンテナ、ネットワーク、ボリューム）
	$(DOCKER_COMPOSE_DEV) down -v

# ビルド
.PHONY: build
build: ## 本番用Dockerイメージをビルド
	$(DOCKER_COMPOSE_DEV) build ocr-api

.PHONY: build-dev
build-dev: ## 開発用Dockerイメージをビルド
	$(DOCKER_COMPOSE_DEV) build $(SERVICE_NAME)

# テスト
.PHONY: test
test: ## テストを実行
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE)

.PHONY: test-unit
test-unit: ## 単体テストのみ実行
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE) go test -v -short -tags=stub ./...

.PHONY: test-integration
test-integration: ## 統合テストを実行
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE) go test -v -run Integration ./...

.PHONY: test-coverage
test-coverage: ## テストカバレッジを生成
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE) go test -v -coverprofile=coverage.out ./...
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE) go tool cover -html=coverage.out -o coverage.html

# コード品質
.PHONY: lint
lint: ## コードリンティングを実行
	$(DOCKER_COMPOSE_DEV) run --rm $(SERVICE_NAME) go vet ./...
	$(DOCKER_COMPOSE_DEV) run --rm $(SERVICE_NAME) go fmt ./...

.PHONY: tidy
tidy: ## go mod tidyを実行
	$(DOCKER_COMPOSE_DEV) run --rm $(SERVICE_NAME) go mod tidy

# 本番環境
.PHONY: prod
prod: ## 本番環境を起動
	$(DOCKER_COMPOSE_DEV) up --build -d ocr-api

.PHONY: prod-logs
prod-logs: ## 本番環境のログを表示
	$(DOCKER_COMPOSE_DEV) logs -f ocr-api

.PHONY: prod-stop
prod-stop: ## 本番環境を停止
	$(DOCKER_COMPOSE_DEV) stop ocr-api

# データベース・外部サービス
.PHONY: services
services: ## 外部サービス（Tesseract等）を起動
	$(DOCKER_COMPOSE_DEV) up -d tesseract

.PHONY: services-stop
services-stop: ## 外部サービスを停止
	$(DOCKER_COMPOSE_DEV) stop tesseract

# デバッグ・ユーティリティ
.PHONY: shell
shell: ## 開発コンテナのシェルに接続
	$(DOCKER_COMPOSE_DEV) exec $(SERVICE_NAME) /bin/bash

.PHONY: shell-test
shell-test: ## テスト環境のシェルに接続
	$(DOCKER_COMPOSE_DEV) run --rm $(TEST_SERVICE) /bin/bash

.PHONY: clean
clean: ## 全てのコンテナ、イメージ、ボリュームを削除
	$(DOCKER_COMPOSE_DEV) down -v --rmi all
	docker system prune -f

.PHONY: reset
reset: clean build-dev ## 環境をリセットして再構築

# APIテスト
.PHONY: api-test
api-test: ## APIエンドポイントをテスト
	@echo "ヘルスチェック..."
	curl -s http://localhost:8080/health | jq .
	@echo "\nサポートされている文書タイプ..."
	curl -s http://localhost:8080/document-types | jq .
	@echo "\nOCRテスト（日本運転免許証）..."
	curl -s -X POST http://localhost:8080/ocr \
		-H "Content-Type: application/json" \
		-d '{"image":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==","documentType":"drivers_license_jp"}' | jq .

# ドキュメント
.PHONY: docs
docs: ## API仕様書を生成（開発用）
	@echo "API仕様書を生成中..."
	@echo "ベースURL: http://localhost:8080"
	@echo "利用可能なエンドポイント:"
	@echo "  GET  /health         - ヘルスチェック"
	@echo "  GET  /document-types - サポートされている文書タイプ"
	@echo "  POST /ocr           - OCR処理"

# 本番デプロイ
.PHONY: deploy
deploy: build ## 本番環境にデプロイ
	@echo "本番環境デプロイを開始..."
	$(DOCKER_COMPOSE_DEV) up -d ocr-api
	@echo "デプロイ完了!"

# ログ監視
.PHONY: logs
logs: ## 全サービスのログを表示
	$(DOCKER_COMPOSE_DEV) logs -f

.PHONY: health
health: ## サービスのヘルスチェック
	@echo "ヘルスチェック実行中..."
	@if curl -f -s http://localhost:8080/health > /dev/null; then \
		echo "✅ サービスは正常に動作しています"; \
	else \
		echo "❌ サービスが応答しません"; \
		exit 1; \
	fi
