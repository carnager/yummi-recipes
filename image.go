package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func downloadImage(dataDir, imageURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	filename := fmt.Sprintf("%x%s", hash[:8], ext)

	uploadDir := filepath.Join(dataDir, "uploads")
	os.MkdirAll(uploadDir, 0755)

	filePath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filename, nil
}

func saveImageBytes(dataDir string, data []byte, originalName string) (string, error) {
	ext := ".jpg"
	lower := strings.ToLower(originalName)
	if strings.HasSuffix(lower, ".png") {
		ext = ".png"
	} else if strings.HasSuffix(lower, ".webp") {
		ext = ".webp"
	}

	hash := sha256.Sum256(data)
	filename := fmt.Sprintf("%x%s", hash[:8], ext)

	uploadDir := filepath.Join(dataDir, "uploads")
	os.MkdirAll(uploadDir, 0755)

	filePath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filename, nil
}

func deleteImage(dataDir, imagePath string) {
	if imagePath == "" {
		return
	}
	fullPath := filepath.Join(dataDir, "uploads", imagePath)
	os.Remove(fullPath)
}
