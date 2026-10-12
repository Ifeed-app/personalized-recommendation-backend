package grpc

import (
	recommendationv1 "backend-diplomski/proto/gen/recommendation/v1"
	"backend-diplomski/services/recommendation-service/internal/client/content"
	"backend-diplomski/services/recommendation-service/internal/client/interaction"
	"backend-diplomski/services/recommendation-service/internal/client/ml"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
)

type Server struct {
	recommendationv1.UnimplementedRecommendationServiceServer

	interactionClient *interaction.Client
	contentClient     *content.Client
	mlClient          *ml.Client
}

func NewServer(
	interactionClient *interaction.Client,
	contentClient *content.Client,
	mlClient *ml.Client,
) *Server {
	return &Server{
		interactionClient: interactionClient,
		contentClient:     contentClient,
		mlClient:          mlClient,
	}
}

type categoryHistoryStats struct {
	count         float64
	positive      float64
	negative      float64
	likes         float64
	unlikes       float64
	skips         float64
	watchRatioSum float64
}

func (s *Server) GetRecommendations(
	ctx context.Context,
	req *recommendationv1.GetRecommendationsRequest,
) (*recommendationv1.GetRecommendationsResponse, error) {

	if req.GetUserId() == "" {
		return nil, fmt.Errorf("user id is required")
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}

	history, err := s.interactionClient.GetUserInteractionHistory(
		ctx,
		req.GetUserId(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get interaction history: %w",
			err,
		)
	}

	if len(history.GetInteractions()) == 0 {
		return s.getColdStartRecommendations(
			ctx,
			limit,
		)
	}

	seenVideos := make(map[string]struct{})
	categoryHistory := make(map[string]*categoryHistoryStats)
	candidateByVideoID := make(map[string]ml.Candidate)

	for _, item := range history.GetInteractions() {

		videoID := item.GetVideoId()

		seenVideos[videoID] = struct{}{}

		metadata, err := s.contentClient.GetVideoMetadata(
			ctx,
			videoID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get metadata for video %s: %w",
				videoID,
				err,
			)
		}

		duration := float64(metadata.GetDurationSecond())
		watchRatio := 0.0
		if duration > 0 {
			watchRatio = float64(item.GetTotalWatchedSeconds()) / duration
		}
		if watchRatio < 0 {
			watchRatio = 0
		}

		if watchRatio > 1 {
			watchRatio = 1
		}
		positive := false
		negative := false

		if item.GetLiked() {
			positive = true
		} else if item.GetUnliked() ||
			item.GetSkipped() {
			negative = true
		} else if watchRatio >= 0.8 {
			positive = true
		} else if watchRatio < 0.2 {
			negative = true
		}

		for _, category := range metadata.GetCategories() {
			categoryName := category.GetName()
			stats, exists := categoryHistory[categoryName]
			if !exists {
				stats = &categoryHistoryStats{}
				categoryHistory[categoryName] = stats
			}
			stats.count++

			if positive {
				stats.positive++
			}

			if negative {
				stats.negative++
			}

			if item.GetLiked() {
				stats.likes++
			}

			if item.GetUnliked() {
				stats.unlikes++
			}

			if item.GetSkipped() {
				stats.skips++
			}

			stats.watchRatioSum += watchRatio
		}
	}

	if len(categoryHistory) == 0 {
		return s.getColdStartRecommendations(
			ctx,
			limit,
		)
	}

	log.Printf(
		"===== USER PREFERENCE SUMMARY: %s =====",
		req.GetUserId(),
	)

	for category, stats := range categoryHistory {
		positiveRate := 0.0
		averageWatchRatio := 0.0

		if stats.count > 0 {
			positiveRate = stats.positive / stats.count
			averageWatchRatio = stats.watchRatioSum / stats.count
		}

		log.Printf(
			"category=%s interactions=%.0f positive=%.0f negative=%.0f likes=%.0f unlikes=%.0f skips=%.0f positive_rate=%.2f avg_watch_ratio=%.2f",
			category,
			stats.count,
			stats.positive,
			stats.negative,
			stats.likes,
			stats.unlikes,
			stats.skips,
			positiveRate,
			averageWatchRatio,
		)
	}

	candidates, err := s.contentClient.ListRecommendationCandidates(
		ctx,
		100)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get recommendation candidates: %w", err)
	}
	mlCandidates := make([]ml.Candidate, 0, len(candidates.GetVideos()))

	for _, candidate := range candidates.GetVideos() {

		videoID := candidate.GetVideoId()

		if _, seen := seenVideos[videoID]; seen {
			continue
		}

		if len(candidate.GetCategories()) == 0 {
			continue
		}

		var historyCount float64
		var historyPositive float64
		var historyNegative float64
		var historyLikes float64
		var historyUnlikes float64
		var historySkips float64
		var historyWatchRatioSum float64

		for _, category := range candidate.GetCategories() {
			categoryName := category.GetName()
			stats := categoryHistory[categoryName]
			if stats == nil {
				continue
			}
			historyCount += stats.count
			historyPositive += stats.positive
			historyNegative += stats.negative
			historyLikes += stats.likes
			historyUnlikes += stats.unlikes
			historySkips += stats.skips
			historyWatchRatioSum += stats.watchRatioSum
		}

		positiveRate := 0.0
		historyWatchRatio := 0.0

		if historyCount > 0 {
			positiveRate =
				historyPositive / historyCount

			historyWatchRatio =
				historyWatchRatioSum / historyCount
		}

		categories := make([]string, 0, len(candidate.GetCategories()))

		for _, category := range candidate.GetCategories() {
			categories = append(categories, category.GetName())
		}

		mlCandidate := ml.Candidate{
			VideoID:                     candidate.GetVideoId(),
			DurationSeconds:             float64(candidate.GetDurationSecond()),
			HistoryCategoryCount:        historyCount,
			HistoryCategoryPositive:     historyPositive,
			HistoryCategoryNegative:     historyNegative,
			HistoryCategoryLikes:        historyLikes,
			HistoryCategoryUnlikes:      historyUnlikes,
			HistoryCategorySkips:        historySkips,
			HistoryCategoryPositiveRate: positiveRate,
			HistoryCategoryWatchRatio:   historyWatchRatio,
			Categories:                  strings.Join(categories, ","),
		}

		mlCandidates = append(mlCandidates, mlCandidate)
		candidateByVideoID[mlCandidate.VideoID] = mlCandidate
	}
	if len(mlCandidates) == 0 {
		return s.getColdStartRecommendations(
			ctx, limit)

	}

	scoredCandidates, err := s.mlClient.ScoreCandidates(ctx, mlCandidates)
	if err != nil {
		return nil, fmt.Errorf("failed to score recommendation candidates: %w")
	}
	sort.Slice(scoredCandidates,
		func(i, j int) bool {
			return scoredCandidates[i].Score >
				scoredCandidates[j].Score
		},
	)

	log.Printf(
		"===== TOP %d RECOMMENDATIONS FOR USER %s =====",
		limit,
		req.GetUserId(),
	)

	for i, scored := range scoredCandidates {
		if i >= int(limit) {
			break
		}

		candidate, exists := candidateByVideoID[scored.VideoID]
		if !exists {
			continue
		}

		log.Printf(
			"#%02d video=%s categories=%s score=%.4f history=%.0f positive=%.0f negative=%.0f likes=%.0f unlikes=%.0f skips=%.0f positive_rate=%.2f watch_ratio=%.2f",
			i+1,
			scored.VideoID,
			candidate.Categories,
			scored.Score,
			candidate.HistoryCategoryCount,
			candidate.HistoryCategoryPositive,
			candidate.HistoryCategoryNegative,
			candidate.HistoryCategoryLikes,
			candidate.HistoryCategoryUnlikes,
			candidate.HistoryCategorySkips,
			candidate.HistoryCategoryPositiveRate,
			candidate.HistoryCategoryWatchRatio,
		)
	}

	if len(scoredCandidates) > 0 {
		middle := len(scoredCandidates) / 2

		log.Printf(
			"ML score distribution: candidates=%d max=%.10f median=%.10f min=%.10f",
			len(scoredCandidates),
			scoredCandidates[0].Score,
			scoredCandidates[middle].Score,
			scoredCandidates[len(scoredCandidates)-1].Score,
		)
	}

	response := &recommendationv1.GetRecommendationsResponse{
		Videos: make([]*recommendationv1.RecommendedVideo, 0, limit),
	}
	for _, candidate := range scoredCandidates {
		if int32(len(response.Videos)) >= limit {
			break
		}
		response.Videos = append(
			response.Videos,
			&recommendationv1.RecommendedVideo{
				VideoId: candidate.VideoID,
				Score:   candidate.Score,
			})
	}

	log.Printf(
		"generated %d ML recommendations for user %s",
		len(response.Videos),
		req.GetUserId(),
	)

	return response, nil
}

func (s *Server) getColdStartRecommendations(
	ctx context.Context,
	limit int32,
) (*recommendationv1.GetRecommendationsResponse, error) {

	candidates, err := s.contentClient.ListRecommendationCandidates(
		ctx,
		100,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get cold start candidates: %w",
			err,
		)
	}

	response := &recommendationv1.GetRecommendationsResponse{
		Videos: make(
			[]*recommendationv1.RecommendedVideo,
			0,
			limit,
		),
	}

	for _, candidate := range candidates.GetVideos() {

		if len(candidate.GetCategories()) == 0 {
			continue
		}

		response.Videos = append(
			response.Videos,
			&recommendationv1.RecommendedVideo{
				VideoId: candidate.GetVideoId(),
				Score:   0,
			},
		)

		if int32(len(response.Videos)) >= limit {
			break
		}
	}

	return response, nil
}
