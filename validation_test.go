package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRequestValidationAndErrorHandling tests the comprehensive request validation
// and error handling functionality without requiring full OCR handler initialization
func TestRequestValidationAndErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		request        OCRRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid request with PNG",
			request: OCRRequest{
				Image:        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
				DocumentType: "drivers_license_jp",
			},
			expectedStatus: http.StatusOK, // Would be OK if processing succeeded
		},
		{
			name: "missing image field",
			request: OCRRequest{
				Image:        "",
				DocumentType: "drivers_license_jp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "image field is required",
		},
		{
			name: "missing document type field",
			request: OCRRequest{
				Image:        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
				DocumentType: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "documentType field is required",
		},
		{
			name: "unsupported document type",
			request: OCRRequest{
				Image:        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
				DocumentType: "invalid_document_type",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "unsupported document type",
		},
		{
			name: "invalid base64 encoding",
			request: OCRRequest{
				Image:        "invalid-base64-data!!!",
				DocumentType: "drivers_license_jp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid base64 encoding",
		},
		{
			name: "unsupported image format",
			request: OCRRequest{
				Image:        "dGhpcyBpcyBub3QgYW4gaW1hZ2U=", // "this is not an image" in base64
				DocumentType: "drivers_license_jp",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "unsupported image format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			err := tt.request.Validate()

			if tt.expectedStatus == http.StatusOK {
				// For valid requests, validation should pass
				if err != nil {
					t.Errorf("Expected validation to pass, but got error: %v", err)
				}
			} else {
				// For invalid requests, validation should fail
				if err == nil {
					t.Errorf("Expected validation to fail, but it passed")
					return
				}

				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

// TestJSONPayloadValidation tests JSON payload parsing and validation
func TestJSONPayloadValidation(t *testing.T) {
	tests := []struct {
		name           string
		jsonPayload    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid JSON payload",
			jsonPayload:    `{"image":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==","documentType":"drivers_license_jp"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON syntax",
			jsonPayload:    `{"image": "test", "documentType": "drivers_license_jp", "invalid": }`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON format",
		},
		{
			name:           "empty JSON payload",
			jsonPayload:    `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "image field is required",
		},
		{
			name:           "malformed JSON",
			jsonPayload:    `{"image": "test"`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that only does JSON parsing and validation
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				var req OCRRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					errorResponse := NewErrorResponse(http.StatusBadRequest, "Invalid JSON format: "+err.Error())
					json.NewEncoder(w).Encode(errorResponse)
					return
				}

				// Validate request
				if err := req.Validate(); err != nil {
					statusCode := http.StatusBadRequest
					if strings.Contains(err.Error(), "unsupported document type") ||
						strings.Contains(err.Error(), "unsupported image format") ||
						strings.Contains(err.Error(), "image size exceeds maximum limit") {
						statusCode = http.StatusUnprocessableEntity
					}

					w.WriteHeader(statusCode)
					errorResponse := NewErrorResponse(statusCode, err.Error())
					json.NewEncoder(w).Encode(errorResponse)
					return
				}

				// If we get here, validation passed
				w.WriteHeader(http.StatusOK)
				response := OCRResponse{
					DocumentType: req.DocumentType,
					Data:         map[string]string{},
				}
				json.NewEncoder(w).Encode(response)
			})

			req, err := http.NewRequest("POST", "/ocr", bytes.NewBufferString(tt.jsonPayload))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedError != "" {
				var errorResponse ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}

				if !strings.Contains(errorResponse.Error.Message, tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errorResponse.Error.Message)
				}
			}
		})
	}
}

// TestHTTPMethodValidation tests HTTP method validation
func TestHTTPMethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST method allowed",
			method:         "POST",
			expectedStatus: http.StatusOK, // Would be OK if processing succeeded
		},
		{
			name:           "GET method not allowed",
			method:         "GET",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "OPTIONS method for CORS",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					w.WriteHeader(http.StatusMethodNotAllowed)
					errorResponse := NewErrorResponse(http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
					json.NewEncoder(w).Encode(errorResponse)
					return
				}

				// For POST requests, return OK (would normally process the request)
				w.WriteHeader(http.StatusOK)
				response := map[string]string{"status": "ok"}
				json.NewEncoder(w).Encode(response)
			})

			var req *http.Request
			var err error

			if tt.method == "POST" {
				validJSON := `{"image":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==","documentType":"drivers_license_jp"}`
				req, err = http.NewRequest(tt.method, "/ocr", bytes.NewBufferString(validJSON))
			} else {
				req, err = http.NewRequest(tt.method, "/ocr", nil)
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check CORS headers for OPTIONS request
			if tt.method == "OPTIONS" {
				expectedHeaders := map[string]string{
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST, OPTIONS",
					"Access-Control-Allow-Headers": "Content-Type",
				}

				for header, expectedValue := range expectedHeaders {
					if actualValue := rr.Header().Get(header); actualValue != expectedValue {
						t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// TestErrorResponseFormat tests that error responses are properly formatted
func TestErrorResponseFormat(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		message      string
		expectedCode int
	}{
		{
			name:         "400 Bad Request error",
			statusCode:   http.StatusBadRequest,
			message:      "Invalid request data",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "422 Unprocessable Entity error",
			statusCode:   http.StatusUnprocessableEntity,
			message:      "Unsupported document type",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			message:      "Internal server error",
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the NewErrorResponse function
			errorResponse := NewErrorResponse(tt.statusCode, tt.message)

			if errorResponse.Error.Code != tt.expectedCode {
				t.Errorf("Expected error code %d, got %d", tt.expectedCode, errorResponse.Error.Code)
			}

			if errorResponse.Error.Message != tt.message {
				t.Errorf("Expected error message '%s', got '%s'", tt.message, errorResponse.Error.Message)
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(errorResponse)
			if err != nil {
				t.Fatalf("Failed to marshal error response: %v", err)
			}

			// Test JSON deserialization
			var deserializedResponse ErrorResponse
			if err := json.Unmarshal(jsonData, &deserializedResponse); err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if deserializedResponse.Error.Code != tt.expectedCode {
				t.Errorf("After JSON round-trip, expected error code %d, got %d", tt.expectedCode, deserializedResponse.Error.Code)
			}

			if deserializedResponse.Error.Message != tt.message {
				t.Errorf("After JSON round-trip, expected error message '%s', got '%s'", tt.message, deserializedResponse.Error.Message)
			}
		})
	}
}
