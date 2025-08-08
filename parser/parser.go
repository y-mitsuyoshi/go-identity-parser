// Package parser provides interfaces and implementations for document parsing
package parser

import "ocr-web-api/imageprocessor"

// DocumentParser defines the interface for parsing different document types
// Uses imageprocessor.Mat for processed image data as specified in the design document
type DocumentParser interface {
	Parse(mat imageprocessor.Mat) (map[string]string, error)
}

// ParserFactory manages document parsers and provides parser selection
type ParserFactory struct {
	parsers map[string]DocumentParser
}

// NewParserFactory creates a new parser factory instance
func NewParserFactory() *ParserFactory {
	factory := &ParserFactory{
		parsers: make(map[string]DocumentParser),
	}

	// Register available parsers
	factory.RegisterParser("drivers_license_jp", NewJPDriverLicenseParser())
	factory.RegisterParser("individual_number_card_jp", NewIndividualNumberCardParser())

	return factory
}

// RegisterParser registers a parser for a specific document type
func (pf *ParserFactory) RegisterParser(documentType string, parser DocumentParser) {
	pf.parsers[documentType] = parser
}

// GetParser returns the appropriate parser for the given document type
func (pf *ParserFactory) GetParser(documentType string) (DocumentParser, error) {
	parser, exists := pf.parsers[documentType]
	if !exists {
		return nil, &UnsupportedDocumentTypeError{DocumentType: documentType}
	}
	return parser, nil
}

// GetSupportedDocumentTypes returns a list of supported document types
func (pf *ParserFactory) GetSupportedDocumentTypes() []string {
	types := make([]string, 0, len(pf.parsers))
	for docType := range pf.parsers {
		types = append(types, docType)
	}
	return types
}

// UnsupportedDocumentTypeError represents an error for unsupported document types
type UnsupportedDocumentTypeError struct {
	DocumentType string
}

func (e *UnsupportedDocumentTypeError) Error() string {
	return "サポートされていない文書タイプです: " + e.DocumentType
}
