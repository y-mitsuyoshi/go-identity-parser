package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// OCRRequest represents the incoming request structure for OCR processing
type OCRRequest struct {
	Image        string `json:"image"`        // Base64 encoded image data
	DocumentType string `json:"documentType"` // Document type identifier
}

// OCRResponse represents the response structure after OCR processing
type OCRResponse struct {
	DocumentType string            `json:"documentType"` // Document type that was processed
	Data         map[string]string `json:"data"`         // Extracted field data
}

// APIError represents error information in API responses
type APIError struct {
	Code    int    `json:"code"`    // HTTP status code
	Message string `json:"message"` // Error message
}

// ErrorResponse represents the complete error response structure
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// Supported document types
const (
	DocumentTypeDriversLicenseJP      = "drivers_license_jp"
	DocumentTypeIndividualNumberCard  = "individual_number_card_jp"
)

// Maximum image size in bytes (10MB)
const MaxImageSize = 10 * 1024 * 1024

// Validate validates the OCR request data
func (req *OCRRequest) Validate() error {
	// Check if required fields are present
	if strings.TrimSpace(req.Image) == "" {
		return errors.New("image field is required")
	}
	
	if strings.TrimSpace(req.DocumentType) == "" {
		return errors.New("documentType field is required")
	}
	
	// Validate document type
	if !isValidDocumentType(req.DocumentType) {
		return fmt.Errorf("unsupported document type: %s", req.DocumentType)
	}
	
	// Validate base64 image data
	if err := validateBase64Image(req.Image); err != nil {
		return err
	}
	
	return nil
}

// isValidDocumentType checks if the document type is supported
func isValidDocumentType(docType string) bool {
	switch docType {
	case DocumentTypeDriversLicenseJP, DocumentTypeIndividualNumberCard:
		return true
	default:
		return false
	}
}

// validateBase64Image validates the base64 encoded image data
func validateBase64Image(imageData string) error {
	// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
	if strings.Contains(imageData, ",") {
		parts := strings.Split(imageData, ",")
		if len(parts) > 1 {
			imageData = parts[1]
		}
	}
	
	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return errors.New("invalid base64 encoding")
	}
	
	// Check image size limit (10MB)
	if len(decodedData) > MaxImageSize {
		return fmt.Errorf("image size exceeds maximum limit of %d bytes", MaxImageSize)
	}
	
	// Check if it's a valid image format (PNG or JPEG)
	if !isValidImageFormat(decodedData) {
		return errors.New("unsupported image format, only PNG and JPEG are supported")
	}
	
	return nil
}

// isValidImageFormat checks if the image data is in PNG or JPEG format
func isValidImageFormat(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	
	// Check PNG signature (89 50 4E 47)
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	
	// Check JPEG signature (FF D8 FF)
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	
	return false
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	}
}
