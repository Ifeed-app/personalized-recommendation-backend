package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DownloadedVideo struct {
	Data        []byte
	ContentType string
	SizeBytes   int64
}

type Downloader struct {
	httpClient *http.Client
	maxBytes   int64
}

func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
		maxBytes: 200 * 1024 * 1024,
	}
}

func (d *Downloader) Download(ctx context.Context, videoURL string) (*DownloadedVideo, error) {
	if videoURL == "" {
		return nil, fmt.Errorf("URL Video is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, videoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create  download req: %w", err)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download video : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	limitReader := io.LimitReader(resp.Body, d.maxBytes+1)
	data, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read video body: %w", err)
	}
	if int64(len(data)) > d.maxBytes {
		return nil, fmt.Errorf("video too large, max allowed size is %d bytess ", d.maxBytes)
	}

	return &DownloadedVideo{
		Data:        data,
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
	}, nil
}
