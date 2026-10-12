package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Candidate struct {
	VideoID                     string  `json:"video_id"`
	DurationSeconds             float64 `json:"duration_seconds"`
	HistoryCategoryCount        float64 `json:"history_category_count"`
	HistoryCategoryPositive     float64 `json:"history_category_positive"`
	HistoryCategoryNegative     float64 `json:"history_category_negative"`
	HistoryCategoryLikes        float64 `json:"history_category_likes"`
	HistoryCategoryUnlikes      float64 `json:"history_category_unlikes"`
	HistoryCategorySkips        float64 `json:"history_category_skips"`
	HistoryCategoryPositiveRate float64 `json:"history_category_positive_rate"`
	HistoryCategoryWatchRatio   float64 `json:"history_category_watch_ratio"`
	Categories                  string  `json:"categories"`
}

type ScoreRequest struct {
	Candidates []Candidate `json:"candidates"`
}

type ScoredCandidate struct {
	VideoID string  `json:"video_id"`
	Score   float64 `json:"score"`
}

type ScoreResponse struct {
	Scores []ScoredCandidate `json:"scores"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) ScoreCandidates(
	ctx context.Context,
	candidates []Candidate,
) ([]ScoredCandidate, error) {

	requestBody := ScoreRequest{
		Candidates: candidates,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal ML score request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/score",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create ML score request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to call ML inference service: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ML inference service returned status %d",
			resp.StatusCode,
		)
	}

	var scoreResponse ScoreResponse

	if err := json.NewDecoder(resp.Body).Decode(
		&scoreResponse,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode ML score response: %w",
			err,
		)
	}

	return scoreResponse.Scores, nil
}
