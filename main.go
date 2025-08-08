package main

import (
	"net/http"
	"os"
)

func main() {
	// Initialize logger
	AppLogger.Info("Starting OCR Web API server...")

	// Initialize OCR handler
	ocrHandler := NewOCRHandler()
	AppLogger.Info("OCR handler initialized successfully")

	// Set up HTTP routes
	http.HandleFunc("/ocr", ocrHandler.HandleOCR)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		HealthHandler(w, r)
	})
	http.HandleFunc("/document-types", ocrHandler.DocumentTypesHandler)
	AppLogger.Info("HTTP routes configured")

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	AppLogger.Infof("OCR Web API server starting on port %s...", port)
	AppLogger.Info("Available endpoints:")
	AppLogger.Info("  POST /ocr - Process OCR requests")
	AppLogger.Info("  GET  /health - Health check")
	AppLogger.Info("  GET  /document-types - Get supported document types")

	AppLogger.Infof("Server ready to accept connections on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		AppLogger.Errorf("Server failed to start: %v", err)
		os.Exit(1)
	}
}

// handleOCR is now replaced by the handler in handler.go
// This function is kept for backward compatibility but is no longer used
func handleOCR(w http.ResponseWriter, r *http.Request) {
	// This implementation is now moved to handler.go
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": {"code": 501, "message": "Please use the new handler implementation"}}`))
}
