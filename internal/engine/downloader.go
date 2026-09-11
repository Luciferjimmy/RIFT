package engine

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Downloader handles securely and reliably fetching files from remote URLs.
type Downloader struct {
	Client  *http.Client
	Retries int
}

// NewDownloader creates a new Downloader instance with a robust transport:
// dial/header timeouts (so a stalled connection fails fast) but no total
// body timeout (large engine tarballs can legitimately take many minutes).
func NewDownloader() *Downloader {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout: 15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
	return &Downloader{
		Client:  &http.Client{Transport: transport}, // follows redirects by default
		Retries: 3,
	}
}

// DownloadFile safely downloads a file to a destination path.
// It downloads to a temporary file first and renames on success to prevent
// corruption. Network/HTTP failures are retried up to Retries times.
func (d *Downloader) DownloadFile(url string, destPath string, makeExecutable bool) error {
	return d.DownloadFileProgress(url, destPath, makeExecutable, nil)
}

// DownloadFileProgress is DownloadFile with a progress callback. The callback
// receives (downloadedBytes, totalBytes); totalBytes is -1 when the server
// does not report a Content-Length. It is invoked after every chunk read.
func (d *Downloader) DownloadFileProgress(url string, destPath string, makeExecutable bool, onProgress func(downloaded, total int64)) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpPath := destPath + ".tmp"
	var lastErr error

	for attempt := 0; attempt <= d.Retries; attempt++ {
		lastErr = d.downloadOnce(url, tmpPath, destPath, makeExecutable, onProgress)
		if lastErr == nil {
			return nil
		}
		fmt.Printf("[Downloader] attempt %d/%d failed: %v\n", attempt+1, d.Retries+1, lastErr)
		if attempt < d.Retries {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}

	os.Remove(tmpPath)
	return lastErr
}

func (d *Downloader) downloadOnce(url, tmpPath, destPath string, makeExecutable bool, onProgress func(downloaded, total int64)) error {
	resp, err := d.Client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to GET url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Reject empty downloads before they shadow a real failure.
	if resp.ContentLength == 0 {
		return fmt.Errorf("empty response body from %s", url)
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()
	defer os.Remove(tmpPath) // Clean up temp file on failure

	var src io.Reader = resp.Body
	if onProgress != nil {
		src = &progressReader{r: resp.Body, total: resp.ContentLength, onProgress: onProgress}
	}

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Close before renaming / chmod so the rename is atomic on macOS.
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Verify the size matches what the server promised (when it told us).
	if resp.ContentLength > 0 {
		if fi, err := os.Stat(tmpPath); err == nil && fi.Size() != resp.ContentLength {
			return fmt.Errorf("downloaded size mismatch: got %d, expected %d", fi.Size(), resp.ContentLength)
		}
	}

	if makeExecutable {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move temp file to destination: %w", err)
	}
	return nil
}

// progressReader wraps an io.Reader and reports bytes read via onProgress.
type progressReader struct {
	r         io.Reader
	total     int64
	read      int64
	onProgress func(downloaded, total int64)
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.read += int64(n)
	if p.onProgress != nil {
		p.onProgress(p.read, p.total)
	}
	return n, err
}
