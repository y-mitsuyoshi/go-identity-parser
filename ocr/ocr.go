package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RegionInfo represents a detected text region with OCR confidence
type RegionInfo struct {
	Text       string
	Confidence float64
	X, Y, W, H int
	Category   string // "name", "address", "date", "number", etc.
}

// OCREngine handles text extraction from images using Tesseract with OpenCV preprocessing
type OCREngine struct {
	tempDir string
}

// NewOCREngine creates a new OCR engine instance
func NewOCREngine() *OCREngine {
	return &OCREngine{
		tempDir: "/tmp",
	}
}

// ExtractText extracts text from image data using Tesseract OCR with OpenCV preprocessing
func (e *OCREngine) ExtractText(imageData []byte) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("cannot process empty image")
	}

	// Preprocess image with OpenCV for better OCR results
	preprocessedImage, err := e.preprocessImageWithOpenCV(imageData)
	if err != nil {
		// If OpenCV preprocessing fails, use original image
		fmt.Printf("Warning: OpenCV preprocessing failed, using original image: %v\n", err)
		preprocessedImage = imageData
	}

	// Create temporary file for the preprocessed image
	tempImageFile, err := os.CreateTemp(e.tempDir, "ocr_preprocessed_*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary image file: %w", err)
	}
	defer os.Remove(tempImageFile.Name())
	defer tempImageFile.Close()

	// Write preprocessed image data to temporary file
	if _, err := tempImageFile.Write(preprocessedImage); err != nil {
		return "", fmt.Errorf("failed to write preprocessed image data to temporary file: %w", err)
	}
	tempImageFile.Close()

	// Create temporary output file path (without extension)
	outputBase := filepath.Join(e.tempDir, "ocr_output_"+filepath.Base(tempImageFile.Name()))
	outputFile := outputBase + ".txt"
	defer os.Remove(outputFile)

	// Run Tesseract OCR with optimized configuration for Japanese documents
	cmd := exec.Command("tesseract", tempImageFile.Name(), outputBase,
		"-l", "jpn+eng",
		"--oem", "1", // Use LSTM OCR Engine Mode only
		"--psm", "3", // Fully automatic page segmentation, but no OSD
		"-c", "tessedit_char_whitelist=0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzあいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをんアイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン一二三四五六七八九十百千万億兆京",
		"--dpi", "300",
	)

	// Set environment to ensure proper operation
	cmd.Env = append(os.Environ(),
		"TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/",
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract OCR command failed: %w", err)
	}

	// Read the OCR output
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output file: %w", err)
	}

	// Clean up the extracted text
	text := strings.TrimSpace(string(outputData))
	if text == "" {
		return "", fmt.Errorf("no text could be extracted from the image")
	}

	return text, nil
}

// ExtractRegions extracts text regions with positional information using OpenCV and Tesseract
func (e *OCREngine) ExtractRegions(imageData []byte) ([]RegionInfo, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("cannot process empty image")
	}

	// Preprocess image with OpenCV
	preprocessedImage, err := e.preprocessImageWithOpenCV(imageData)
	if err != nil {
		// If OpenCV preprocessing fails, use original image
		fmt.Printf("Warning: OpenCV preprocessing failed, using original image: %v\n", err)
		preprocessedImage = imageData
	}

	// Create temporary file for the preprocessed image
	tempImageFile, err := os.CreateTemp(e.tempDir, "ocr_regions_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary image file: %w", err)
	}
	defer os.Remove(tempImageFile.Name())
	defer tempImageFile.Close()

	// Write preprocessed image data to temporary file
	if _, err := tempImageFile.Write(preprocessedImage); err != nil {
		return nil, fmt.Errorf("failed to write preprocessed image data to temporary file: %w", err)
	}
	tempImageFile.Close()

	// Use Tesseract to extract text with bounding box information
	outputBase := filepath.Join(e.tempDir, "ocr_regions_"+filepath.Base(tempImageFile.Name()))
	tsvFile := outputBase + ".tsv"
	defer os.Remove(tsvFile)

	// Run Tesseract with TSV output for bounding boxes - optimized for Japanese
	cmd := exec.Command("tesseract", tempImageFile.Name(), outputBase,
		"-l", "jpn+eng",
		"--oem", "1", // Use LSTM OCR Engine Mode only
		"--psm", "3", // Fully automatic page segmentation, but no OSD
		"-c", "tessedit_char_whitelist=0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzあいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをんアイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン一二三四五六七八九十百千万億兆京",
		"--dpi", "300",
		"tsv",
	)

	cmd.Env = append(os.Environ(),
		"TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/",
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tesseract TSV command failed: %w", err)
	}

	// Parse TSV output to extract regions
	regions, err := e.parseTesseractTSV(tsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Tesseract TSV output: %w", err)
	}

	return regions, nil
}

// preprocessImageWithOpenCV applies OpenCV preprocessing to improve OCR accuracy
func (e *OCREngine) preprocessImageWithOpenCV(imageData []byte) ([]byte, error) {
	// Create temporary files for OpenCV processing
	inputFile, err := os.CreateTemp(e.tempDir, "opencv_input_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create input file: %w", err)
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	outputFile, err := os.CreateTemp(e.tempDir, "opencv_output_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	defer outputFile.Close()

	// Write input image
	if _, err := inputFile.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write input image: %w", err)
	}
	inputFile.Close()
	outputFile.Close()

	// Create Python script for enhanced OpenCV preprocessing optimized for Japanese text
	pythonScript := fmt.Sprintf(`
import cv2
import numpy as np
import sys

try:
    # Read the image
    img = cv2.imread('%s')
    if img is None:
        print("Error: Could not load image", file=sys.stderr)
        sys.exit(1)

    # Resize image to improve OCR accuracy (minimum 300 DPI equivalent)
    height, width = img.shape[:2]
    if height < 600 or width < 800:
        scale_factor = max(600/height, 800/width)
        new_width = int(width * scale_factor)
        new_height = int(height * scale_factor)
        img = cv2.resize(img, (new_width, new_height), interpolation=cv2.INTER_CUBIC)

    # Convert to grayscale
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Apply CLAHE (Contrast Limited Adaptive Histogram Equalization) for better contrast
    clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(8,8))
    enhanced = clahe.apply(gray)

    # Apply bilateral filter to reduce noise while preserving edges
    filtered = cv2.bilateralFilter(enhanced, 9, 75, 75)

    # Apply adaptive threshold optimized for Japanese characters
    thresh = cv2.adaptiveThreshold(filtered, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY, 15, 4)

    # Morphological operations to connect broken characters (common in Japanese text)
    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (1, 1))
    cleaned = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel)
    
    # Remove small noise
    kernel2 = cv2.getStructuringElement(cv2.MORPH_RECT, (2, 2))
    cleaned = cv2.morphologyEx(cleaned, cv2.MORPH_OPEN, kernel2)

    # Final enhancement for better character recognition
    final = cv2.medianBlur(cleaned, 3)

    # Save the processed image
    success = cv2.imwrite('%s', final)
    if not success:
        print("Error: Could not save processed image", file=sys.stderr)
        sys.exit(1)
        
    print("Enhanced image preprocessing completed successfully")

except Exception as e:
    print(f"Error during image processing: {e}", file=sys.stderr)
    sys.exit(1)
`, inputFile.Name(), outputFile.Name())

	// Write Python script to temporary file
	scriptFile, err := os.CreateTemp(e.tempDir, "opencv_script_*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create script file: %w", err)
	}
	defer os.Remove(scriptFile.Name())
	defer scriptFile.Close()

	if _, err := scriptFile.WriteString(pythonScript); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	scriptFile.Close()

	// Execute Python script
	cmd := exec.Command("python3", scriptFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("OpenCV preprocessing failed: %w, output: %s", err, string(output))
	}

	// Read the processed image
	processedData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read processed image: %w", err)
	}

	return processedData, nil
}

// parseTesseractTSV parses Tesseract TSV output to extract text regions
func (e *OCREngine) parseTesseractTSV(tsvFile string) ([]RegionInfo, error) {
	data, err := os.ReadFile(tsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read TSV file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var regions []RegionInfo

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue // Skip malformed lines
		}

		// Extract relevant fields from TSV
		text := strings.TrimSpace(fields[11])
		if text == "" {
			continue
		}

		// Parse coordinates and confidence (simplified)
		region := RegionInfo{
			Text:       text,
			Confidence: 0.8, // Default confidence, could be parsed from TSV
			Category:   e.categorizeText(text),
		}

		regions = append(regions, region)
	}

	return regions, nil
}

// categorizeText attempts to categorize extracted text
func (e *OCREngine) categorizeText(text string) string {
	// Simple heuristics for categorizing text
	if strings.Contains(text, "年") && strings.Contains(text, "月") {
		return "date"
	}
	if strings.Contains(text, "都") || strings.Contains(text, "県") || strings.Contains(text, "市") || strings.Contains(text, "区") {
		return "address"
	}
	if len(text) >= 2 && len(text) <= 10 && !strings.ContainsAny(text, "0123456789") {
		return "name"
	}
	if strings.ContainsAny(text, "0123456789") && len(text) >= 4 {
		return "number"
	}
	return "other"
}

// Close cleans up resources (no-op for this implementation)
func (e *OCREngine) Close() error {
	return nil
}
