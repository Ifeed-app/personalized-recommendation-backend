package pexels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type SearchParams struct {
	Query              string
	Orientation        string
	Limit              int
	Page               int
	MinDurationSeconds int
	MaxDurationSeconds int
}

type ExternalVideo struct {
	Source       string
	ExternalID   string
	Title        string
	Description  string
	VideoURL     string
	ThumbnailURL string
	PageURL      string
	Author       string
	Query        string
	DurationSec  int
	Width        int
	Height       int
	Orientation  string
}

type pexelsSearchResponse struct {
	Page         int           `json:"page"`
	PerPage      int           `json:"per_page"`
	TotalResults int           `json:"total_results"`
	URL          string        `json:"url"`
	Videos       []pexelsVideo `json:"videos"`
}

type pexelsVideo struct {
	ID            int               `json:"id"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Duration      int               `json:"duration"`
	URL           string            `json:"url"`
	Image         string            `json:"image"`
	User          pexelsUser        `json:"user"`
	VideoFiles    []pexelsVideoFile `json:"video_files"`
	VideoPictures []pexelsPicture   `json:"video_pictures"`
}

type pexelsUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type pexelsVideoFile struct {
	ID       int     `json:"id"`
	Quality  string  `json:"quality"`
	FileType string  `json:"file_type"`
	Width    *int    `json:"width"`
	Height   *int    `json:"height"`
	FPS      float64 `json:"fps"`
	Link     string  `json:"link"`
}

type pexelsPicture struct {
	ID      int    `json:"id"`
	Picture string `json:"picture"`
	NR      int    `json:"nr"`
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SearchVideos(ctx context.Context, params SearchParams) ([]ExternalVideo, error) {
	params = normalizeParams(params)

	endpoint, err := url.Parse(c.baseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("invalid Pexels API URL: %w", err)
	}

	q := endpoint.Query()
	q.Set("query", params.Query)
	q.Set("orientation", params.Orientation)
	q.Set("per_page", fmt.Sprintf("%d", params.Limit))
	q.Set("page", fmt.Sprintf("%d", params.Page))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pexels request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Pexels API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("pexels API returned status %d", resp.StatusCode)
	}

	var searchResponse pexelsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Pexels response: %w", err)
	}
	videos := make([]ExternalVideo, 0)
	for _, video := range searchResponse.Videos {
		if !isValidShortVideo(video, params.MinDurationSeconds, params.MaxDurationSeconds) {
			continue
		}

		bestFile := chooseBestMP4(video.VideoFiles)
		if bestFile == nil {
			continue
		}
		videos = append(videos, ExternalVideo{
			Source:       "pexels",
			ExternalID:   fmt.Sprintf("%d", video.ID),
			Title:        params.Query,
			Description:  "Imported from Pexels",
			VideoURL:     bestFile.Link,
			ThumbnailURL: video.Image,
			PageURL:      video.URL,
			Author:       video.User.Name,
			Query:        params.Query,
			DurationSec:  video.Duration,
			Width:        getDimension(bestFile.Width),
			Height:       getDimension(bestFile.Height),
			Orientation:  params.Orientation,
		})
	}
	return videos, nil
}

func normalizeParams(params SearchParams) SearchParams {
	if params.Limit <= 0 {
		params.Limit = 10
	}

	if params.Limit > 80 {
		params.Limit = 80
	}

	if params.Page <= 0 {
		params.Page = 1
	}

	if params.Orientation == "" {
		params.Orientation = "portrait"
	}

	if params.MinDurationSeconds <= 0 {
		params.MinDurationSeconds = 3
	}

	if params.MaxDurationSeconds <= 0 {
		params.MaxDurationSeconds = 120
	}

	return params
}

func isValidShortVideo(video pexelsVideo, minDuration int, maxDuration int) bool {
	if video.Duration < minDuration {
		return false
	}

	if video.Duration > maxDuration {
		return false
	}

	if video.Height <= video.Width {
		return false
	}

	return true
}

func chooseBestMP4(files []pexelsVideoFile) *pexelsVideoFile {
	const (
		targetWidth  = 720
		targetHeight = 1280
		maxWidth     = 1080
		maxHeight    = 1920
	)
	var best *pexelsVideoFile
	bestScore := int(^uint(0) >> 1)

	for i := range files {
		file := &files[i]

		if file.FileType != "video/mp4" {
			continue
		}

		width := getDimension(file.Width)
		height := getDimension(file.Height)

		if width <= 0 || height <= 0 {
			continue
		}

		if width >= height {
			continue
		}

		if width > maxWidth || height > maxHeight {
			continue
		}
		score := abs(width-targetWidth) + abs(height-targetHeight)

		if best == nil || score < bestScore {
			best = file
			bestScore = score
		}

	}

	return best
}

func getDimension(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value

}
