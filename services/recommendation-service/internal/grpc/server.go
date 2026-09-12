package grpc

import (
	recommendationv1 "backend-diplomski/proto/gen/recommendation/v1"
	"backend-diplomski/services/recommendation-service/internal/client/content"
	"backend-diplomski/services/recommendation-service/internal/client/interaction"
	"backend-diplomski/services/recommendation-service/internal/recommendation"
	"context"
	"fmt"
	"log"
	"sort"
)

type Server struct {
	recommendationv1.UnimplementedRecommendationServiceServer

	interactionClient *interaction.Client
	contentClient     *content.Client
}

func NewServer(
	interactionClient *interaction.Client,
	contentClient *content.Client,
) *Server {
	return &Server{
		interactionClient: interactionClient,
		contentClient:     contentClient,
	}
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
	categoryScores := make(map[string]float64)

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

		interactionScore := recommendation.CalculateInteractionScore(
			recommendation.InteractionScoreInput{
				Liked:               item.GetLiked(),
				Unliked:             item.GetUnliked(),
				Skipped:             item.GetSkipped(),
				TotalWatchedSeconds: item.GetTotalWatchedSeconds(),
				DurationSeconds:     metadata.GetDurationSecond(),
			},
		)

		for _, category := range metadata.GetCategories() {
			categoryScores[category.GetName()] += interactionScore
		}
	}

	if len(categoryScores) == 0 {
		return s.getColdStartRecommendations(
			ctx,
			limit,
		)
	}

	log.Printf(
		"user %s category preference profile: %+v",
		req.GetUserId(),
		categoryScores,
	)

	candidates, err := s.contentClient.ListRecommendationCandidates(
		ctx,
		100,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get recommendation candidates: %w",
			err,
		)
	}

	type scoredCandidate struct {
		videoID string
		score   float64
	}

	scoredCandidates := make(
		[]scoredCandidate,
		0,
		len(candidates.GetVideos()),
	)

	for _, candidate := range candidates.GetVideos() {

		videoID := candidate.GetVideoId()

		if _, seen := seenVideos[videoID]; seen {
			continue
		}

		if len(candidate.GetCategories()) == 0 {
			continue
		}

		categoryNames := make(
			[]string,
			0,
			len(candidate.GetCategories()),
		)

		for _, category := range candidate.GetCategories() {
			categoryNames = append(
				categoryNames,
				category.GetName(),
			)
		}

		score := recommendation.CalculateCandidateScore(
			categoryScores,
			categoryNames,
		)

		scoredCandidates = append(
			scoredCandidates,
			scoredCandidate{
				videoID: videoID,
				score:   score,
			},
		)
	}

	sort.Slice(
		scoredCandidates,
		func(i, j int) bool {
			return scoredCandidates[i].score >
				scoredCandidates[j].score
		},
	)

	response := &recommendationv1.GetRecommendationsResponse{
		Videos: make(
			[]*recommendationv1.RecommendedVideo,
			0,
			limit,
		),
	}

	for _, candidate := range scoredCandidates {

		if int32(len(response.Videos)) >= limit {
			break
		}

		response.Videos = append(
			response.Videos,
			&recommendationv1.RecommendedVideo{
				VideoId: candidate.videoID,
				Score:   candidate.score,
			},
		)
	}

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
