package parser

import (
	"fmt"
	"ocr-web-api/imageprocessor"
	"ocr-web-api/ocr"
)

// extractTextUsingOCR performs OCR text extraction from the image
func (p *IndividualNumberCardParser) extractTextUsingOCR(mat imageprocessor.Mat) (string, error) {

	if len(mat) == 0 {
		return "", fmt.Errorf("cannot process empty image")
	}

	// Initialize the OCR engine
	engine := ocr.NewOCREngine()
	defer engine.Close()

	// Extract text using the engine
	text, err := engine.ExtractText([]byte(mat))
	if err != nil {
		return "", fmt.Errorf("OCR engine failed to extract text: %w", err)
	}

	return text, nil
}
