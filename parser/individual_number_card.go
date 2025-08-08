package parser

import (
	"fmt"
	"ocr-web-api/imageprocessor"
	"ocr-web-api/ocr"
	"regexp"
	"strings"
)

// IndividualNumberCardParser handles parsing of Japanese Individual Number Card documents
type IndividualNumberCardParser struct {
	patterns map[string]*regexp.Regexp
}

// NewIndividualNumberCardParser creates a new Individual Number Card parser instance
func NewIndividualNumberCardParser() *IndividualNumberCardParser {
	return &IndividualNumberCardParser{
		patterns: initIndividualNumberCardPatterns(),
	}
}

// Parse extracts structured data from an Individual Number Card image
func (p *IndividualNumberCardParser) Parse(mat imageprocessor.Mat) (map[string]string, error) {
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
		return nil, fmt.Errorf("validation failed for individual number card data: %w", err)
	}

	return extractedData, nil
}

// parseWithRegionDetection uses OpenCV region detection for more accurate field extraction
func (p *IndividualNumberCardParser) parseWithRegionDetection(mat imageprocessor.Mat) (map[string]string, error) {
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

	// Process regions based on category and content
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
				} else {
					extractedData["expiry_date"] = region.Text
				}
			}
		case "number":
			if isValidIndividualNumber(region.Text) {
				extractedData["individual_number"] = region.Text
			}
		}
	}

	// Also try to extract municipality and name using enhanced patterns
	if municipality := extractMunicipalityFromRegions(regions); municipality != "" {
		extractedData["municipality"] = municipality
	}

	if name := extractNameFromRegions(regions); name != "" {
		extractedData["name"] = name
	}

	return extractedData, nil
}

// Validation helper functions
func isValidName(text string) bool {
	if len(text) < 2 || len(text) > 20 {
		return false
	}
	// Check if text contains only valid Japanese characters for names
	return !strings.ContainsAny(text, "0123456789年月日都道府県市区町村")
}

func isValidAddress(text string) bool {
	if len(text) < 5 {
		return false
	}
	// Check if text contains address indicators
	return strings.ContainsAny(text, "都道府県市区町村")
}

func isValidDate(text string) bool {
	// Check for Japanese date format
	return strings.Contains(text, "年") && strings.Contains(text, "月")
}

func isValidIndividualNumber(text string) bool {
	if len(text) != 12 {
		return false
	}
	// Check if all characters are digits
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Enhanced extraction functions
func extractMunicipalityFromRegions(regions []ocr.RegionInfo) string {
	for _, region := range regions {
		if strings.Contains(region.Text, "都") || strings.Contains(region.Text, "県") ||
			strings.Contains(region.Text, "市") || strings.Contains(region.Text, "区") {
			// Clean up the municipality text
			municipality := strings.TrimSpace(region.Text)
			if len(municipality) >= 3 && len(municipality) <= 20 {
				return municipality
			}
		}
	}
	return ""
}

func extractNameFromRegions(regions []ocr.RegionInfo) string {
	for _, region := range regions {
		if region.Category == "name" || (len(region.Text) >= 2 && len(region.Text) <= 10 &&
			!strings.ContainsAny(region.Text, "0123456789年月日都道府県市区町村個人番号")) {
			name := strings.TrimSpace(region.Text)
			if isValidName(name) {
				return name
			}
		}
	}
	return ""
}

// parseTextWithRegex extracts structured data from OCR text using regex patterns
func (p *IndividualNumberCardParser) parseTextWithRegex(ocrText string) (map[string]string, error) {
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

	if _, exists := extractedData["individual_number"]; !exists {
		if altPattern, exists := p.patterns["individual_number_alt"]; exists {
			matches := altPattern.FindStringSubmatch(ocrText)
			if len(matches) > 1 {
				extractedData["individual_number"] = strings.TrimSpace(matches[1])
			}
		}
	}

	// Post-process extracted data
	p.postProcessExtractedData(extractedData)

	return extractedData, nil
}

// postProcessExtractedData cleans and normalizes extracted data
func (p *IndividualNumberCardParser) postProcessExtractedData(data map[string]string) {
	// Normalize individual number format
	if individualNumber, exists := data["individual_number"]; exists {
		// Ensure proper format XXXX-XXXX-XXXX
		cleaned := strings.ReplaceAll(individualNumber, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		if len(cleaned) == 12 {
			// Format as XXXX-XXXX-XXXX
			formatted := cleaned[:4] + "-" + cleaned[4:8] + "-" + cleaned[8:12]
			data["individual_number"] = formatted
		}
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

	// Normalize gender field
	if gender, exists := data["gender"]; exists {
		cleaned := strings.TrimSpace(gender)
		if cleaned == "男性" || cleaned == "男" {
			data["gender"] = "男"
		} else if cleaned == "女性" || cleaned == "女" {
			data["gender"] = "女"
		}
	}
}

// initIndividualNumberCardPatterns initializes regex patterns for Individual Number Card fields
func initIndividualNumberCardPatterns() map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)

	// Name pattern (氏名)
	patterns["name"] = regexp.MustCompile(`(?:氏名|氏\s*名)\s*[:：]?\s*([^\r\n]+)`)

	// Address pattern (住所)
	patterns["address"] = regexp.MustCompile(`(?:住所|住\s*所)\s*[:：]?\s*([^\r\n]+)`)

	// Birth date pattern (生年月日)
	patterns["birth_date"] = regexp.MustCompile(`(?:生年月日|生\s*年\s*月\s*日)\s*[:：]?\s*([^\r\n]+)`)

	// Gender pattern (性別)
	patterns["gender"] = regexp.MustCompile(`(?:性別|性\s*別)\s*[:：]?\s*([男女])`)

	// Individual number pattern (個人番号) - 12 digit number
	patterns["individual_number"] = regexp.MustCompile(`(?:個人番号|個人\s*番号)\s*[:：]?\s*(\d{4}\s*\d{4}\s*\d{4}|\d{12})`)

	// Issue date pattern (交付年月日)
	patterns["issue_date"] = regexp.MustCompile(`(?:交付年月日|交付\s*年\s*月\s*日)\s*[:：]?\s*([^\r\n]+)`)

	// Expiry date pattern (有効期限)
	patterns["expiry_date"] = regexp.MustCompile(`(?:有効期限|有効\s*期\s*限)\s*[:：]?\s*([^\r\n]+)`)

	// Alternative name pattern (カタカナ or 漢字) - Fixed Unicode ranges
	patterns["name_alt"] = regexp.MustCompile(`([ァ-ヴー一-龯ひ-ゖ]+\s+[ァ-ヴー一-龯ひ-ゖ]+)`)

	// Date patterns with specific Japanese date formats
	patterns["birth_date_alt"] = regexp.MustCompile(`([平成|昭和|令和]\d{1,2}年\d{1,2}月\d{1,2}日|\d{4}年\d{1,2}月\d{1,2}日)`)

	// Individual number with specific format (XXXX-XXXX-XXXX)
	patterns["individual_number_alt"] = regexp.MustCompile(`(\d{4}-\d{4}-\d{4})`)

	// Municipality pattern (市区町村)
	patterns["municipality"] = regexp.MustCompile(`([^都道府県]+[都道府県][^市区町村]+[市区町村])`)

	return patterns
}

// validateExtractedData validates the extracted data for required fields
func (p *IndividualNumberCardParser) validateExtractedData(data map[string]string) error {
	requiredFields := []string{"name"}
	fmt.Println("sss")
	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists || strings.TrimSpace(value) == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Additional validation for specific fields
	if individualNumber, exists := data["individual_number"]; exists {
		// Remove spaces and hyphens, then validate length
		cleanNumber := strings.ReplaceAll(individualNumber, " ", "")
		cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
		if len(cleanNumber) != 12 {
			return fmt.Errorf("invalid individual number format: expected 12 digits, got %d", len(cleanNumber))
		}
	}

	// Gender validation
	if gender, exists := data["gender"]; exists {
		if gender != "男" && gender != "女" {
			return fmt.Errorf("invalid gender value: expected '男' or '女', got '%s'", gender)
		}
	}

	return nil
}
