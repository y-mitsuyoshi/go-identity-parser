#!/bin/bash

# OCR APIリクエスト用のJSONファイルを作成するスクリプト

if [ $# -ne 2 ]; then
    echo "Usage: $0 <base64_file> <output_json_file>"
    echo "Example: $0 tmp/mynumber_card_base64.txt tmp/ocr_request.json"
    exit 1
fi

BASE64_FILE="$1"
OUTPUT_JSON="$2"

if [ ! -f "$BASE64_FILE" ]; then
    echo "Error: Base64 file '$BASE64_FILE' not found"
    exit 1
fi

echo "Creating OCR API request JSON..."

# Base64データを読み取り
BASE64_DATA=$(cat "$BASE64_FILE")

# JSONファイルを作成
cat > "$OUTPUT_JSON" << EOF
{
  "image": "$BASE64_DATA",
  "document_type": "individual_number_card",
  "language": "jpn"
}
EOF

echo "OCR request JSON created: $OUTPUT_JSON"
echo "JSON file size: $(wc -c < "$OUTPUT_JSON") bytes"
