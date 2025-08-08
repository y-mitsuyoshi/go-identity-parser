package ocr

// Engine defines the interface for OCR operations
type Engine interface {
	ExtractText(imageData []byte) (string, error)
	Close() error
}
