package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloaderDownloadAndVerify(t *testing.T) {
	content := []byte("VirtIO and Windows ISO test binary content stream 1234567890")
	hash := sha256.Sum256(content)
	expectedSHA256 := hex.EncodeToString(hash[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "60")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "winqemu-dl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	destFile := filepath.Join(tmpDir, "test.iso")
	dl := NewDownloader()

	var lastProgress Progress
	err = dl.Download(context.Background(), server.URL, destFile, expectedSHA256, func(p Progress) {
		lastProgress = p
	})
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if lastProgress.BytesDownloaded != int64(len(content)) {
		t.Errorf("expected %d bytes downloaded, got %d", len(content), lastProgress.BytesDownloaded)
	}

	saved, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(saved) != string(content) {
		t.Errorf("content mismatch")
	}
}

func TestDownloaderChecksumMismatch(t *testing.T) {
	content := []byte("corrupted payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "winqemu-dl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	destFile := filepath.Join(tmpDir, "bad.iso")
	dl := NewDownloader()

	err = dl.Download(context.Background(), server.URL, destFile, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", nil)
	if err == nil {
		t.Errorf("expected checksum mismatch error, got nil")
	}
}

func TestDownloaderRejectsHTML(t *testing.T) {
	htmlContent := []byte("<!DOCTYPE html><html><body>Error 404 Not Found</body></html>")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(htmlContent)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "winqemu-dl-html-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	destFile := filepath.Join(tmpDir, "dummy.iso")
	dl := NewDownloader()

	err = dl.Download(context.Background(), server.URL, destFile, "", nil)
	if err == nil {
		t.Errorf("expected error when server returns HTML, got nil")
	}
}

func TestDownloaderRejectsSmallFile(t *testing.T) {
	content := []byte("too short")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "winqemu-dl-small-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	destFile := filepath.Join(tmpDir, "small.iso")
	dl := NewDownloader()
	dl.MinBytes = 1024 * 1024 // 1 MB minimum

	err = dl.Download(context.Background(), server.URL, destFile, "", nil)
	if err == nil {
		t.Errorf("expected error when file is smaller than MinBytes, got nil")
	}
}
