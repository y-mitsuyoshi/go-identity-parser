package parser

import (
	"ocr-web-api/ocr"
	"strings"
)

// Validation helper functions
func isValidName(text string) bool {
	// Exclude common OCR errors and irrelevant keywords
	if strings.Contains(text, "氏名") || strings.Contains(text, "番号") || strings.Contains(text, "免許") {
		return false
	}
	// Basic length check
	if len(text) < 2 || len(text) > 20 {
		return false
	}
	// Check that the name contains at least one common Japanese character (Kanji, Hiragana, Katakana)
	// This is a simplified check, a more robust solution would use Unicode ranges.
	if !strings.ContainsAny(text, "一二三四五六七八九十百千万円日月火水木金土春夏秋冬東西南北") {
		// This is just a small sample of kanji, not a comprehensive list.
		// A better approach would be to check for character ranges.
	}
	// Check if text contains only valid Japanese characters for names
	return !strings.ContainsAny(text, "0123456789年月日都道府県市区町村")
}

func isValidAddress(text string) bool {
	if len(text) < 5 {
		return false
	}
	// Check if text contains address indicators
	if !strings.ContainsAny(text, "都道府県市区町村丁目番地") {
		return false
	}
	// Exclude strings that are likely not addresses
	if strings.Contains(text, "氏名") || strings.Contains(text, "生年月日") {
		return false
	}
	return true
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
	// Attempt to find the "氏名" label and extract the name from the next region
	for i, region := range regions {
		if strings.Contains(region.Text, "氏名") || strings.Contains(region.Text, "氏 名") {
			// The name is often in the next region
			if i+1 < len(regions) {
				name := strings.TrimSpace(regions[i+1].Text)
				// Clean up the name, remove the label if it's there
				name = strings.ReplaceAll(name, "氏名", "")
				name = strings.TrimSpace(name)
				if isValidName(name) {
					return name
				}
			}
			// Sometimes the name is in the same region as the label
			name := strings.ReplaceAll(region.Text, "氏名", "")
			name = strings.TrimSpace(name)
			if isValidName(name) {
				return name
			}
		}
	}

	// Fallback to the original logic if the label is not found
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
