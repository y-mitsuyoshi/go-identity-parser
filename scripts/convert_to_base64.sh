#!/bin/bash

# 画像ファイルをBase64に変換するスクリプト

if [ $# -ne 2 ]; then
    echo "Usage: $0 <input_image> <output_base64_file>"
    echo "Example: $0 tmp/マイナンバーカード表.jpg tmp/mynumber_card_base64.txt"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    exit 1
fi

echo "Converting $INPUT_FILE to Base64..."

# Base64エンコード（改行なし）
base64 -w 0 "$INPUT_FILE" > "$OUTPUT_FILE"

echo "Base64 encoded data saved to: $OUTPUT_FILE"
echo "File size: $(wc -c < "$OUTPUT_FILE") bytes"
