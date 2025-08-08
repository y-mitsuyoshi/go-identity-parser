# Go Identity Parser - OCR Web API

多様な身分証明書の画像から構造化されたデータを抽出する、拡張性の高いREST APIをGo言語で実装しています。このアプリケーションはDockerコンテナで動作し、OCR技術を活用して画像から必要な情報を自動抽出します。

## 特徴

- **多文書対応**: 日本の運転免許証、個人番号カードに対応
- **拡張可能アーキテクチャ**: Strategyパターンによる新しい文書タイプの簡単追加
- **Docker対応**: マルチステージビルドによる最適化されたコンテナ
- **高精度OCR**: OpenCVによる画像前処理とTesseractによる文字認識
- **REST API**: 標準的なHTTPインターフェース
- **構造化ログ**: レベル別ログ出力

## サポートされている文書タイプ

### 日本運転免許証 (`drivers_license_jp`)
抽出可能フィールド:
- `name`: 氏名
- `address`: 住所
- `birth_date`: 生年月日
- `license_number`: 免許証番号
- `issue_date`: 交付年月日
- `expiry_date`: 有効期限
- `license_class`: 免許の種類

### 個人番号カード (`individual_number_card_jp`)
抽出可能フィールド:
- `name`: 氏名
- `address`: 住所
- `birth_date`: 生年月日
- `gender`: 性別
- `individual_number`: 個人番号
- `issue_date`: 交付年月日
- `expiry_date`: 有効期限

## API エンドポイント

### POST /ocr
身分証明書の画像からデータを抽出します。

**リクエスト:**
```json
{
  "image": "base64_encoded_image_data",
  "documentType": "drivers_license_jp"
}
```

**レスポンス:**
```json
{
  "documentType": "drivers_license_jp",
  "data": {
    "name": "田中太郎",
    "address": "東京都港区赤坂1-2-3",
    "birth_date": "平成5年12月25日",
    "license_number": "1234 5678 9012",
    "issue_date": "令和5年1月15日",
    "expiry_date": "令和10年12月25日"
  }
}
```

### GET /health
アプリケーションのヘルスチェックを行います。

**レスポンス:**
```json
{
  "status": "healthy",
  "service": "OCR Web API",
  "version": "1.0.0"
}
```

### GET /document-types
サポートされている文書タイプの一覧を取得します。

**レスポンス:**
```json
{
  "supported_document_types": [
    "drivers_license_jp",
    "individual_number_card_jp"
  ],
  "total_count": 2
}
```

## セットアップ

### 前提条件

- Docker
- Docker Compose
- Git

### 開発環境でのセットアップ

このプロジェクトはDocker環境での開発を推奨しています。すべてのビルド、テスト、実行はDockerコンテナ内で行われ、ホストマシンに依存関係をインストールする必要はありません。

#### 1. プロジェクトのクローン

```bash
git clone https://github.com/y-mitsuyoshi/go-identity-parser.git
cd go-identity-parser
```

#### 2. 開発環境の起動

```bash
# 開発環境を起動（ホットリロード付き）
make dev

# または直接docker composeを使用
docker compose up --build ocr-api-dev
```

#### 3. 利用可能なMakeコマンド

開発を効率化するためのMakefileコマンドが用意されています：

```bash
# ヘルプを表示
make help

# 開発環境
make dev          # 開発環境を起動（ホットリロード付き）
make dev-bg       # バックグラウンドで開発環境を起動
make dev-logs     # 開発環境のログを表示
make dev-stop     # 開発環境を停止
make dev-down     # 開発環境を削除

# テスト
make test              # 全テストを実行
make test-unit         # 単体テストのみ実行
make test-integration  # 統合テストを実行
make test-coverage     # テストカバレッジを生成

# コード品質
make lint         # コードリンティングを実行
make tidy         # go mod tidyを実行

# ビルド
make build        # 本番用Dockerイメージをビルド
make build-dev    # 開発用Dockerイメージをビルド

# ユーティリティ
make shell        # 開発コンテナのシェルに接続
make api-test     # APIエンドポイントをテスト
make health       # サービスのヘルスチェック
make clean        # 全てのコンテナ・イメージを削除
make reset        # 環境をリセットして再構築
```

#### 4. 開発ワークフロー

開発時の基本的なワークフローは以下の通りです：

```bash
# 1. 開発環境を起動
make dev

# 2. 別のターミナルでテストを実行
make test

# 3. コードを編集（ホットリロードで自動反映）
# ファイルを保存すると自動的にサーバーが再起動します

# 4. APIをテスト
make api-test

# 5. 開発完了時
make dev-stop
```

#### 5. 開発コンテナの特徴

- **ホットリロード**: Airを使用してコード変更時に自動でサーバーを再起動
- **ボリュームマウント**: ローカルのソースコードがコンテナにマウントされ、変更が即座に反映
- **Go Modules キャッシュ**: 依存関係のダウンロード時間を短縮
- **デバッグ対応**: 開発モードでのログレベルはDEBUGに設定

#### 6. デバッグとトラブルシューティング

```bash
# 開発コンテナにシェルで接続
make shell

# ログを確認
make dev-logs

# サービスの状態を確認
make health

# 環境をリセット（問題が発生した場合）
make reset
```

### 本番環境でのセットアップ

#### Docker でのセットアップ

1. Dockerイメージをビルド:

```bash
make build
# または
docker build -t go-identity-parser .
```

2. 本番環境を起動:

```bash
make prod
# または
docker compose up -d ocr-api
```

### ローカル開発環境でのセットアップ（非推奨）

Dockerを使用しない場合は以下の手順でセットアップできますが、依存関係の管理が複雑になるため推奨しません：

1. 依存関係をインストール:

```bash
go mod download
```

2. 必要なライブラリをインストール (Ubuntu/Debian):

```bash
sudo apt-get update
sudo apt-get install -y tesseract-ocr tesseract-ocr-jpn tesseract-ocr-eng libopencv-dev
```

3. アプリケーションを実行:

```bash
go run .
```

## 環境変数

- `PORT`: サーバーポート (デフォルト: 8080)
- `LOG_LEVEL`: ログレベル (DEBUG, INFO, WARN, ERROR) (デフォルト: INFO)
- `TESSERACT_DATA_PATH`: Tesseractデータファイルパス

## テスト

### 単体テストの実行

```bash
go test ./...
```

### 統合テストの実行

```bash
go test -v ./...
```

### テストカバレッジの確認

```bash
go test -cover ./...
```

## アーキテクチャ

```text
├── main.go                 # HTTPサーバーエントリーポイント
├── handler.go              # HTTPリクエストハンドラー
├── types.go                # データ型定義
├── logger.go               # ログ機能
├── parser/                 # 文書パーサー
│   ├── parser.go          # インターフェース定義とファクトリー
│   ├── drivers_license_jp.go  # 日本運転免許証パーサー
│   └── individual_number_card.go  # 個人番号カードパーサー
├── imageprocessor/         # 画像前処理
│   ├── processor.go       # OpenCV画像処理
│   ├── base64_decoder.go  # Base64デコーダー
│   └── interface.go       # インターフェース定義
└── ocr/                   # OCRエンジン
    └── ocr.go             # Tesseract OCR操作
```

## 新しい文書タイプの追加

新しい文書タイプを追加するには、以下の手順に従ってください:

1. `parser/` ディレクトリに新しいパーサーファイルを作成
2. `DocumentParser` インターフェースを実装
3. `parser.go` の `NewParserFactory()` 関数でパーサーを登録

例:

```go
// parser/passport_us.go
type USPassportParser struct {
    patterns map[string]*regexp.Regexp
}

func NewUSPassportParser() *USPassportParser {
    return &USPassportParser{
        patterns: initUSPassportPatterns(),
    }
}

func (p *USPassportParser) Parse(mat gocv.Mat) (map[string]string, error) {
    // パーサー実装
}
```

## エラーハンドリング

APIは以下のHTTPステータスコードを返します:

- `200 OK`: 正常処理完了
- `400 Bad Request`: 無効なリクエスト形式
- `405 Method Not Allowed`: サポートされていないHTTPメソッド
- `422 Unprocessable Entity`: 処理できないデータ
- `500 Internal Server Error`: サーバー内部エラー

エラーレスポンス形式:

```json
{
  "error": {
    "code": 400,
    "message": "Invalid JSON format"
  }
}
```

## パフォーマンス考慮事項

- 画像サイズ制限: 最大10MB推奨
- 同時処理: CPU数に基づく制限
- メモリ管理: OpenCVマトリックスの適切な解放
- リクエストタイムアウト: 30秒

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 貢献

プルリクエストや issue の報告を歓迎します。大きな変更を行う前に、まず issue を作成して変更内容について議論してください。

## 開発者

- Yuma Mitsuyoshi ([@y-mitsuyoshi](https://github.com/y-mitsuyoshi))
