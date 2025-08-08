# マルチステージビルド - ビルドステージ
FROM ubuntu:24.04 AS builder

# パッケージリストを更新し、必要なライブラリをインストール
RUN apt-get update && apt-get install -y \
    golang-1.21 \
    git \
    build-essential \
    pkg-config \
    tesseract-ocr \
    tesseract-ocr-jpn \
    tesseract-ocr-eng \
    libtesseract-dev \
    libleptonica-dev \
    && rm -rf /var/lib/apt/lists/*

# Goのパスを設定
ENV PATH="/usr/lib/go-1.21/bin:${PATH}"
ENV GOPATH="/go"
ENV GOROOT="/usr/lib/go-1.21"

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesファイルをコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド（実際のOCR実装を使用）
RUN go build -o ocr-api .

# 実行ステージ
FROM ubuntu:24.04

# パッケージリストを更新し、必要なライブラリをインストール
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-jpn \
    tesseract-ocr-eng \
    python3 \
    python3-pip \
    python3-opencv \
    && rm -rf /var/lib/apt/lists/*

# Python依存関係をインストール
RUN apt-get remove -y python3-numpy && \
    pip3 install --no-cache-dir --break-system-packages opencv-python \
    && rm -rf /var/lib/apt/lists/*

# ビルドステージからバイナリをコピー
COPY --from=builder /app/ocr-api /usr/local/bin/ocr-api

# 実行権限を付与
RUN chmod +x /usr/local/bin/ocr-api

# ポート8080を公開
EXPOSE 8080

# 環境変数のデフォルト値を設定
ENV PORT=8080
ENV LOG_LEVEL=INFO

# アプリケーションを実行
CMD ["ocr-api"]
