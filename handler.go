package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"ocr-web-api/imageprocessor"
	"ocr-web-api/parser"
	"strings"
	"time"
)

// OCRHandler handles OCR API requests
type OCRHandler struct {
	parserFactory  *parser.ParserFactory
	imageProcessor *imageprocessor.ImageProcessor
}

// NewOCRHandler creates a new OCR handler instance
func NewOCRHandler() *OCRHandler {
	return &OCRHandler{
		parserFactory:  parser.NewParserFactory(),
		imageProcessor: imageprocessor.NewImageProcessor(),
	}
}

// HandleOCR processes OCR requests
func (h *OCRHandler) HandleOCR(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		AppLogger.Warnf("Invalid method attempted: %s from %s", r.Method, r.RemoteAddr)
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	// Create request context with 30-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Track request start time for logging
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		AppLogger.Infof("Request from %s completed in %v", r.RemoteAddr, duration)
	}()

	AppLogger.Infof("OCR request received from %s", r.RemoteAddr)

	// Parse request body
	var req OCRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		AppLogger.Errorf("Failed to parse request JSON from %s: %v", r.RemoteAddr, err)
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}

	AppLogger.Debugf("Request parsed: documentType=%s, imageSize=%d bytes", req.DocumentType, len(req.Image))

	// Validate request using the comprehensive validation from types.go
	if err := req.Validate(); err != nil {
		AppLogger.Warnf("Request validation failed from %s: %v", r.RemoteAddr, err)
		// Determine appropriate status code based on error type
		statusCode := h.getErrorStatusCode(err)
		h.sendErrorResponse(w, statusCode, err.Error())
		return
	}

	// Process the OCR request with timeout context
	response, err := h.processOCRRequestWithTimeout(ctx, &req)
	if err != nil {
		// Check if the error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			AppLogger.Errorf("Request timeout for %s from %s after 30 seconds", req.DocumentType, r.RemoteAddr)
			h.sendErrorResponse(w, http.StatusRequestTimeout, "Request timeout: processing exceeded 30 seconds")
			return
		}

		AppLogger.Errorf("OCR processing error for %s from %s: %v", req.DocumentType, r.RemoteAddr, err)
		h.sendErrorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	AppLogger.Infof("OCR processing completed successfully for %s from %s", req.DocumentType, r.RemoteAddr)

	// Send successful response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		AppLogger.Errorf("Failed to encode response for %s: %v", r.RemoteAddr, err)
	}
}

// getErrorStatusCode determines the appropriate HTTP status code based on the error type
func (h *OCRHandler) getErrorStatusCode(err error) int {
	errMsg := err.Error()

	// 400 Bad Request - for missing required fields and invalid base64
	if strings.Contains(errMsg, "field is required") ||
		strings.Contains(errMsg, "invalid base64 encoding") {
		return http.StatusBadRequest
	}

	// 422 Unprocessable Entity - for unsupported document types, image format issues, size limits
	if strings.Contains(errMsg, "unsupported document type") ||
		strings.Contains(errMsg, "unsupported image format") ||
		strings.Contains(errMsg, "image size exceeds maximum limit") {
		return http.StatusUnprocessableEntity
	}

	// Default to 400 Bad Request for other validation errors
	return http.StatusBadRequest
}

// processOCRRequest processes the OCR request and returns extracted data
func (h *OCRHandler) processOCRRequest(req *OCRRequest) (*OCRResponse, error) {
	// Step 1: Process the image (decode Base64, preprocess)
	processedMat, err := h.imageProcessor.ProcessImage(req.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Step 2: Get the appropriate parser for the document type
	parser, err := h.parserFactory.GetParser(req.DocumentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser: %w", err)
	}

	// Step 3: Parse the processed image using the selected parser
	// Pass the processed image data to the parser
	extractedData, err := parser.Parse(processedMat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Step 4: Create and return response
	response := &OCRResponse{
		DocumentType: req.DocumentType,
		Data:         extractedData,
	}

	return response, nil
}

// processOCRRequestWithTimeout processes the OCR request with context timeout
func (h *OCRHandler) processOCRRequestWithTimeout(ctx context.Context, req *OCRRequest) (*OCRResponse, error) {
	// Use a channel to handle the result from the processing
	resultChan := make(chan *OCRResponse, 1)
	errorChan := make(chan error, 1)

	// Run the OCR processing in a goroutine
	go func() {
		response, err := h.processOCRRequest(req)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- response
		}
	}()

	// Wait for either completion or timeout
	select {
	case response := <-resultChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// sendErrorResponse sends an error response in JSON format
func (h *OCRHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errorResponse := ErrorResponse{
		Error: APIError{
			Code:    statusCode,
			Message: message,
		},
	}

	AppLogger.Debugf("Sending error response: %d - %s", statusCode, message)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		AppLogger.Errorf("Failed to encode error response: %v", err)
	}
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		AppLogger.Warnf("Invalid method attempted on health endpoint: %s from %s", r.Method, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	AppLogger.Debugf("Health check requested from %s", r.RemoteAddr)

	healthResponse := map[string]interface{}{
		"status":  "healthy",
		"service": "OCR Web API",
		"version": "1.0.0",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse)
}

// DocumentTypesHandler returns supported document types
func (h *OCRHandler) DocumentTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		AppLogger.Warnf("Invalid method attempted on document-types endpoint: %s from %s", r.Method, r.RemoteAddr)
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET.")
		return
	}

	AppLogger.Debugf("Document types requested from %s", r.RemoteAddr)

	supportedTypes := h.parserFactory.GetSupportedDocumentTypes()

	response := map[string]interface{}{
		"supported_document_types": supportedTypes,
		"total_count":              len(supportedTypes),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
