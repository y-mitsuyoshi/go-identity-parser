package imageprocessor

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Base64Decoder handles base64 image decoding
type Base64Decoder struct{}

// NewBase64Decoder creates a new Base64Decoder instance
func NewBase64Decoder() *Base64Decoder {
	return &Base64Decoder{}
}

// DecodeBase64 decodes a Base64 encoded image string to byte slice with validation
func (d *Base64Decoder) DecodeBase64(base64Image string) ([]byte, error) {
	if base64Image == "" {
		return nil, fmt.Errorf("base64 image string is empty")
	}

	// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
	if strings.Contains(base64Image, "base64,") {
		parts := strings.Split(base64Image, "base64,")
		if len(parts) >= 2 {
			base64Image = parts[1]
		}
	}

	// Decode base64 string
	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	if len(imageData) == 0 {
		return nil, fmt.Errorf("decoded image data is empty")
	}

	return imageData, nil
}
