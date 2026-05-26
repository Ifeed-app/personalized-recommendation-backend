package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"video-worker/internal/message"
	"video-worker/internal/rabbit"

	"github.com/google/uuid"
)

type FetchScheduler struct {
	publisher *rabbit.Publisher

	intervalSeconds int
	queries         []string
	limit           int
	orientation     string
	minDuration     int
	maxDuration     int

	queryPages map[string]int
}

func NewFetchScheduler(
	publisher *rabbit.Publisher,
	intervalSeconds int,
	queries []string,
	limit int,
	orientation string,
	minDuration int,
	maxDuration int,
) *FetchScheduler {
	return &FetchScheduler{
		publisher:       publisher,
		intervalSeconds: intervalSeconds,
		queries:         queries,
		limit:           limit,
		orientation:     orientation,
		minDuration:     minDuration,
		maxDuration:     maxDuration,
		queryPages:      make(map[string]int),
	}
}

func (s *FetchScheduler) Start(ctx context.Context) {
	if s.intervalSeconds <= 0 {
		s.intervalSeconds = 300
	}

	log.Printf("video fetch scheduler started, interval=%ds", s.intervalSeconds)

	s.publishBatch(ctx)

	ticker := time.NewTicker(time.Duration(s.intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("video fetch scheduler stopped")
			return

		case <-ticker.C:
			s.publishBatch(ctx)
		}
	}
}

func (s *FetchScheduler) publishBatch(ctx context.Context) {
	for _, rawQuery := range s.queries {
		query := strings.TrimSpace(rawQuery)
		if query == "" {
			continue
		}
		currentPage := s.queryPages[query]
		if currentPage <= 0 {
			currentPage = 1
		}
		s.queryPages[query] = currentPage + 1
		msg := message.ShortVideoFetchMessage{
			JobID:              fmt.Sprintf("pexels-%s-%s", query, uuid.NewString()),
			Provider:           "pexels",
			Query:              query,
			Orientation:        s.orientation,
			Limit:              s.limit,
			Page:               currentPage,
			MinDurationSeconds: s.minDuration,
			MaxDurationSeconds: s.maxDuration,
		}

		publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := s.publisher.PublishFetchJob(publishCtx, msg)
		cancel()

		if err != nil {
			log.Printf("failed to publish fetch job for query=%s: %v", query, err)
			continue
		}
	}
}
