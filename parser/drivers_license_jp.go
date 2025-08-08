package parser

import (
	"fmt"
	"ocr-web-api/imageprocessor"
	"ocr-web-api/ocr"
	"regexp"
	"strings"
)

// JPDriverLicenseParser handles parsing of Japanese driver's license documents
type JPDriverLicenseParser struct {
	patterns map[string]*regexp.Regexp
}

// NewJPDriverLicenseParser creates a new Japanese driver's license parser instance
func NewJPDriverLicenseParser() *JPDriverLicenseParser {
	return &JPDriverLicenseParser{
		patterns: initJPDriverLicensePatterns(),
	}
}

// Parse extracts structured data from a Japanese driver's license image
func (p *JPDriverLicenseParser) Parse(mat imageprocessor.Mat) (map[string]string, error) {
	// Step 1: Try region-based extraction with OpenCV for better accuracy
	extractedData, err := p.parseWithRegionDetection(mat)
	if err == nil && len(extractedData) > 0 {
		// Step 1.5: Validate the extracted data from region detection
		if validationErr := p.validateExtractedData(extractedData); validationErr == nil {
			return extractedData, nil
		} else {
			fmt.Printf("Region-based extraction validation failed, falling back to full OCR: %v\n", validationErr)
		}
	}

	// Step 2: Fallback to traditional OCR text extraction
	ocrText, err := p.extractTextUsingOCR(mat)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text via OCR: %w", err)
	}

	// Step 3: Parse the text using regex patterns
	extractedData, err = p.parseTextWithRegex(ocrText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text with regex: %w", err)
	}

	// Step 4: Validate the extracted data
	if err := p.validateExtractedData(extractedData); err != nil {
		return nil, fmt.Errorf("validation failed for driver's license data: %w", err)
	}

	return extractedData, nil
}

// parseWithRegionDetection uses OpenCV region detection for more accurate field extraction
func (p *JPDriverLicenseParser) parseWithRegionDetection(mat imageprocessor.Mat) (map[string]string, error) {
	// Convert Mat to image data
	imageData, err := mat.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert Mat to bytes: %w", err)
	}

	// Create OCR engine
	engine := ocr.NewOCREngine()
	defer engine.Close()

	// Extract text regions with positional information
	regions, err := engine.ExtractRegions(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract regions: %w", err)
	}

	extractedData := make(map[string]string)

	// Process regions based on category and content for driver's license
	for _, region := range regions {
		switch region.Category {
		case "name":
			if isValidName(region.Text) {
				extractedData["name"] = region.Text
			}
		case "address":
			if isValidAddress(region.Text) {
				extractedData["address"] = region.Text
			}
		case "date":
			if isValidDate(region.Text) {
				if strings.Contains(region.Text, "生") {
					extractedData["birth_date"] = region.Text
				} else if strings.Contains(region.Text, "交付") {
					extractedData["issue_date"] = region.Text
				} else if strings.Contains(region.Text, "有効") {
					extractedData["expiry_date"] = region.Text
				}
			}
		case "number":
			if isValidLicenseNumber(region.Text) {
				extractedData["license_number"] = region.Text
			}
		}
	}

	// Enhanced extraction for driver's license specific fields
	if municipality := extractMunicipalityFromRegions(regions); municipality != "" {
		extractedData["municipality"] = municipality
	}

	if name := extractNameFromRegions(regions); name != "" {
		extractedData["name"] = name
	}

	return extractedData, nil
}

// parseTextWithRegex extracts structured data from OCR text using regex patterns
func (p *JPDriverLicenseParser) parseTextWithRegex(ocrText string) (map[string]string, error) {
	extractedData := make(map[string]string)

	// Apply each regex pattern to extract relevant fields
	for fieldName, pattern := range p.patterns {
		matches := pattern.FindStringSubmatch(ocrText)
		if len(matches) > 1 {
			// Clean up the extracted text
			value := strings.TrimSpace(matches[1])
			if value != "" {
				extractedData[fieldName] = value
			}
		}
	}

	// If primary patterns don't match, try alternative patterns
	if _, exists := extractedData["name"]; !exists {
		if altPattern, exists := p.patterns["name_alt"]; exists {
			matches := altPattern.FindStringSubmatch(ocrText)
			if len(matches) > 1 {
				extractedData["name"] = strings.TrimSpace(matches[1])
			}
		}
	}

	if _, exists := extractedData["birth_date"]; !exists {
		if altPattern, exists := p.patterns["birth_date_alt"]; exists {
			matches := altPattern.FindStringSubmatch(ocrText)
			if len(matches) > 1 {
				extractedData["birth_date"] = strings.TrimSpace(matches[1])
			}
		}
	}

	// Post-process extracted data
	p.postProcessExtractedData(extractedData)

	return extractedData, nil
}

// postProcessExtractedData cleans and normalizes extracted data
func (p *JPDriverLicenseParser) postProcessExtractedData(data map[string]string) {
	// Normalize license number format
	if licenseNumber, exists := data["license_number"]; exists {
		// Remove extra spaces and normalize format
		cleaned := strings.ReplaceAll(licenseNumber, "  ", " ")
		cleaned = strings.TrimSpace(cleaned)
		data["license_number"] = cleaned
	}

	// Clean up address field
	if address, exists := data["address"]; exists {
		// Remove excess whitespace and newlines
		cleaned := strings.ReplaceAll(address, "\n", " ")
		cleaned = strings.ReplaceAll(cleaned, "\t", " ")
		// Collapse multiple spaces to single space
		for strings.Contains(cleaned, "  ") {
			cleaned = strings.ReplaceAll(cleaned, "  ", " ")
		}
		data["address"] = strings.TrimSpace(cleaned)
	}

	// Normalize name field
	if name, exists := data["name"]; exists {
		// Ensure proper spacing between family and given name
		cleaned := strings.TrimSpace(name)
		// If there's no space between kanji characters, add one
		if !strings.Contains(cleaned, " ") && len([]rune(cleaned)) > 2 {
			runes := []rune(cleaned)
			if len(runes) >= 4 {
				// Insert space after presumed family name (first 2-3 characters)
				familyNameEnd := 2
				if len(runes) > 5 {
					familyNameEnd = 3
				}
				cleaned = string(runes[:familyNameEnd]) + " " + string(runes[familyNameEnd:])
			}
		}
		data["name"] = cleaned
	}
}

// initJPDriverLicensePatterns initializes regex patterns for Japanese driver's license fields
func initJPDriverLicensePatterns() map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)

	// Name pattern (氏名) - 設計書の例に基づく
	patterns["name"] = regexp.MustCompile(`氏名\s*([^\s]+\s+[^\s]+)`)

	// Address pattern (住所) - 設計書の例に基づく
	patterns["address"] = regexp.MustCompile(`住所\s*([^\n]+)`)

	// Birth date pattern (生年月日) - 設計書の例に基づく
	patterns["birth_date"] = regexp.MustCompile(`生年月日\s*([^\s]+)`)

	// License number pattern (免許証番号) - 12 digit number
	patterns["license_number"] = regexp.MustCompile(`(?:免許証番号|免許証\s*番号)\s*[:：]?\s*(\d{4}\s*\d{4}\s*\d{4}|\d{12})`)

	// Issue date pattern (交付年月日)
	patterns["issue_date"] = regexp.MustCompile(`(?:交付年月日|交付\s*年\s*月\s*日)\s*[:：]?\s*([^\r\n]+)`)

	// Expiry date pattern (有効期限)
	patterns["expiry_date"] = regexp.MustCompile(`(?:有効期限|有効\s*期\s*限)\s*[:：]?\s*([^\r\n]+)`)

	// License class pattern (免許の種類)
	patterns["license_class"] = regexp.MustCompile(`(?:免許の種類|免許\s*の\s*種類|種類)\s*[:：]?\s*([^\r\n]+)`)

	// Alternative name pattern (カタカナ or 漢字) - Fixed Unicode ranges
	patterns["name_alt"] = regexp.MustCompile(`([ァ-ヴー一-龯ひ-ゖ]+\s+[ァ-ヴー一-龯ひ-ゖ]+)`)

	// Date patterns with specific Japanese date formats
	patterns["birth_date_alt"] = regexp.MustCompile(`([平成|昭和|令和]\d{1,2}年\d{1,2}月\d{1,2}日|\d{4}年\d{1,2}月\d{1,2}日)`)

	return patterns
}

// validateExtractedData validates the extracted data for required fields
func (p *JPDriverLicenseParser) validateExtractedData(data map[string]string) error {
	requiredFields := []string{"name"}

	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists || strings.TrimSpace(value) == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Additional validation for specific fields
	if licenseNumber, exists := data["license_number"]; exists {
		// Remove spaces and validate length
		cleanNumber := strings.ReplaceAll(licenseNumber, " ", "")
		if len(cleanNumber) != 12 {
			return fmt.Errorf("invalid license number format: expected 12 digits, got %d", len(cleanNumber))
		}
	}

	return nil
}

// Validation helper function for license number
func isValidLicenseNumber(text string) bool {
	// Remove spaces and check if it's 12 digits
	cleaned := strings.ReplaceAll(text, " ", "")
	if len(cleaned) != 12 {
		return false
	}
	// Check if all characters are digits
	for _, r := range cleaned {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
