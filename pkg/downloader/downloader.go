package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// VirtIOStableURL points to the latest verified VirtIO Windows drivers ISO published by Red Hat/Fedora.
	VirtIOStableURL = "https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/stable-virtio/virtio-win.iso"
)

// Progress reports live download stats.
type Progress struct {
	BytesDownloaded int64
	TotalBytes      int64
	Percent         float64
	SpeedBytesSec   float64
	Done            bool
}

// Downloader manages HTTP downloads with resume and verification.
type Downloader struct {
	client *http.Client
}

// NewDownloader creates an instance with custom timeout and transport settings.
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 0, // No hard overall timeout for multi-gigabyte ISO downloads
		},
	}
}

// Download fetches a file from url and saves it to destPath, supporting resume if a partial file exists.
func (d *Downloader) Download(ctx context.Context, url, destPath, expectedSHA256 string, onProgress func(Progress)) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	partPath := destPath + ".part"
	var existingBytes int64 = 0

	fi, err := os.Stat(partPath)
	if err == nil {
		existingBytes = fi.Size()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if existingBytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingBytes))
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	var file *os.File
	var totalBytes int64

	switch resp.StatusCode {
	case http.StatusPartialContent:
		// Resumed
		file, err = os.OpenFile(partPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		totalBytes = existingBytes + resp.ContentLength
	case http.StatusOK:
		// Fresh start or server doesn't support Range
		file, err = os.Create(partPath)
		if err != nil {
			return err
		}
		existingBytes = 0
		totalBytes = resp.ContentLength
	case http.StatusRequestedRangeNotSatisfiable:
		// Already fully downloaded or invalid range
		file, err = os.Create(partPath)
		if err != nil {
			return err
		}
		existingBytes = 0
		totalBytes = 0
	default:
		return fmt.Errorf("server responded with status %d: %s", resp.StatusCode, resp.Status)
	}
	defer file.Close()

	buf := make([]byte, 64*1024)
	downloaded := existingBytes
	startTime := time.Now()
	lastReport := time.Now()
	lastDownloaded := downloaded

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := file.Write(buf[:n]); wErr != nil {
				return fmt.Errorf("disk write failed: %w", wErr)
			}
			downloaded += int64(n)

			if onProgress != nil && time.Since(lastReport) > 200*time.Millisecond {
				now := time.Now()
				durationSec := now.Sub(lastReport).Seconds()
				var speed float64
				if durationSec > 0 {
					speed = float64(downloaded-lastDownloaded) / durationSec
				}

				percent := 0.0
				if totalBytes > 0 {
					percent = (float64(downloaded) / float64(totalBytes)) * 100
				}

				onProgress(Progress{
					BytesDownloaded: downloaded,
					TotalBytes:      totalBytes,
					Percent:         percent,
					SpeedBytesSec:   speed,
					Done:            false,
				})

				lastReport = now
				lastDownloaded = downloaded
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return fmt.Errorf("read error during download: %w", rErr)
		}
	}

	_ = file.Sync()
	_ = file.Close()

	// Verify checksum if supplied
	if expectedSHA256 != "" {
		hash, err := calculateSHA256(partPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		if hash != expectedSHA256 {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, hash)
		}
	}

	// Rename .part to destPath
	if err := os.Rename(partPath, destPath); err != nil {
		return fmt.Errorf("failed to finalize downloaded file: %w", err)
	}

	if onProgress != nil {
		onProgress(Progress{
			BytesDownloaded: downloaded,
			TotalBytes:      totalBytes,
			Percent:         100.0,
			SpeedBytesSec:   float64(downloaded) / (time.Since(startTime).Seconds() + 0.001),
			Done:            true,
		})
	}

	return nil
}

func calculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
