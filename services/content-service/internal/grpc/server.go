package grpc

import (
	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/services/content-service/db/sqlc"
	"backend-diplomski/services/content-service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Server struct {
	contentv1.UnimplementedContentServiceServer
	queries *sqlc.Queries
	storage *storage.MinioStorage
}

func NewServer(queries *sqlc.Queries, storage *storage.MinioStorage) *Server {
	return &Server{
		queries: queries,
		storage: storage,
	}
}

func (s *Server) CreateVideo(
	ctx context.Context,
	req *contentv1.CreateVideoRequest,
) (*contentv1.CreateVideoResponse, error) {

	log.Printf(
		"CREATE VIDEO RECEIVED storageKey=%s categories=%v tags=%v",
		req.GetStroageKey(),
		req.GetCategories(),
		req.GetTags(),
	)

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if req.StroageKey == "" {
		return nil, fmt.Errorf("storage key is required")
	}

	if req.MimeType == "" {
		req.MimeType = "video/mp4"
	}

	if req.Status == "" {
		req.Status = "READY"
	}

	var video sqlc.ContentServiceVideo

	existingVideo, err := s.queries.GetVideoByStorageKey(
		ctx,
		req.StroageKey,
	)

	if err == nil {
		video = existingVideo
	} else {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"failed to check existing video: %w",
				err,
			)
		}

		videoID := uuid.New()

		video, err = s.queries.CreateVideo(
			ctx,
			sqlc.CreateVideoParams{
				ID:              toPgUUID(videoID),
				CreatorID:       nullPgUUID(),
				Title:           req.Title,
				Description:     nullableText(req.Description),
				StorageKey:      req.StroageKey,
				ThumbnailKey:    nullableText(req.ThumbnailKey),
				MimeType:        req.MimeType,
				DurationSeconds: req.DurationSecond,
				Status:          req.Status,
			},
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to create video: %w",
				err,
			)
		}
	}

	for _, rawCategory := range req.GetCategories() {
		categoryName := strings.ToLower(
			strings.TrimSpace(rawCategory),
		)

		if categoryName == "" {
			continue
		}

		category, err := s.queries.GetOrCreateCategory(
			ctx,
			sqlc.GetOrCreateCategoryParams{
				ID:   toPgUUID(uuid.New()),
				Name: categoryName,
			},
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to get or create category %q: %w",
				categoryName,
				err,
			)
		}

		if err := s.queries.AddVideoCategory(
			ctx,
			sqlc.AddVideoCategoryParams{
				VideoID:    video.ID,
				CategoryID: category.ID,
			},
		); err != nil {
			return nil, fmt.Errorf(
				"failed to add category %q to video: %w",
				categoryName,
				err,
			)
		}
	}

	for _, rawTag := range req.GetTags() {
		tagName := strings.ToLower(
			strings.TrimSpace(rawTag),
		)

		if tagName == "" {
			continue
		}

		tag, err := s.queries.GetOrCreateTag(
			ctx,
			sqlc.GetOrCreateTagParams{
				ID:   toPgUUID(uuid.New()),
				Name: tagName,
			},
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to get or create tag %q: %w",
				tagName,
				err,
			)
		}

		if err := s.queries.AddVideoTag(
			ctx,
			sqlc.AddVideoTagParams{
				VideoID: video.ID,
				TagID:   tag.ID,
			},
		); err != nil {
			return nil, fmt.Errorf(
				"failed to add tag %q to video: %w",
				tagName,
				err,
			)
		}
	}

	return &contentv1.CreateVideoResponse{
		Id:         uuid.UUID(video.ID.Bytes).String(),
		Title:      video.Title,
		StroageKey: video.StorageKey,
		Status:     video.Status,
	}, nil
}
func (s *Server) ListFeedVideos(ctx context.Context, req *contentv1.ListFeedVideosRequest) (*contentv1.ListFeedVideosResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	videos, err := s.queries.ListReadyVideos(ctx, sqlc.ListReadyVideosParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ready videos: %w", err)

	}
	response := &contentv1.ListFeedVideosResponse{
		Videos: make([]*contentv1.FeedVideo, 0, len(videos)),
	}
	for _, video := range videos {
		videoURL, err := s.storage.PresignedObject(ctx, video.StorageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create predesigned video url for storage key=%s: %w", err)
		}
		response.Videos = append(response.Videos, &contentv1.FeedVideo{
			Id:             uuid.UUID(video.ID.Bytes).String(),
			Title:          video.Title,
			Description:    textValue(video.Description),
			StroageKey:     video.StorageKey,
			VideoUrl:       videoURL,
			ThumbnailKey:   textValue(video.ThumbnailKey),
			MimeType:       video.MimeType,
			DurationSecond: video.DurationSeconds,
			Status:         video.Status,
			CreatedAt:      video.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	return response, nil
}

func (s *Server) ListRecommendationCandidates(
	ctx context.Context,
	req *contentv1.ListRecommendationCandidatesRequest,
) (*contentv1.ListRecommendationCandidatesResponse, error) {

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 100
	}

	videos, err := s.queries.ListReadyVideos(
		ctx,
		sqlc.ListReadyVideosParams{
			Limit:  limit,
			Offset: 0,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list recommendation candidates: %w",
			err,
		)
	}

	response := &contentv1.ListRecommendationCandidatesResponse{}

	for _, video := range videos {

		categories, err := s.queries.ListCategoriesByVideoID(
			ctx,
			video.ID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to list categories for video %s: %w",
				uuid.UUID(video.ID.Bytes).String(),
				err,
			)
		}

		candidate := &contentv1.RecommendationCandidate{
			VideoId:        uuid.UUID(video.ID.Bytes).String(),
			DurationSecond: video.DurationSeconds,
		}

		for _, category := range categories {
			candidate.Categories = append(
				candidate.Categories,
				&contentv1.Category{
					Id:   uuid.UUID(category.ID.Bytes).String(),
					Name: category.Name,
				},
			)
		}

		response.Videos = append(
			response.Videos,
			candidate,
		)
	}

	return response, nil
}

func (s *Server) GetVideoMetadata(
	ctx context.Context,
	req *contentv1.GetVideoMetadataRequest,
) (*contentv1.GetVideoMetadataResponse, error) {

	videoID, err := uuid.Parse(req.GetVideoId())
	if err != nil {
		return nil, fmt.Errorf("invalid video id: %w", err)
	}

	pgVideoID := toPgUUID(videoID)

	video, err := s.queries.GetVideoByID(
		ctx,
		pgVideoID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get video: %w",
			err,
		)
	}

	categories, err := s.queries.ListCategoriesByVideoID(
		ctx,
		pgVideoID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list categories: %w",
			err,
		)
	}

	tags, err := s.queries.ListTagsByVideoID(
		ctx,
		pgVideoID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list tags: %w",
			err,
		)
	}

	response := &contentv1.GetVideoMetadataResponse{
		VideoId:        req.GetVideoId(),
		DurationSecond: video.DurationSeconds,
	}

	for _, category := range categories {
		response.Categories = append(
			response.Categories,
			&contentv1.Category{
				Id:   uuid.UUID(category.ID.Bytes).String(),
				Name: category.Name,
			},
		)
	}

	for _, tag := range tags {
		response.Tags = append(
			response.Tags,
			&contentv1.Tag{
				Id:   uuid.UUID(tag.ID.Bytes).String(),
				Name: tag.Name,
			},
		)
	}

	return response, nil
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func nullPgUUID() pgtype.UUID {
	return pgtype.UUID{
		Valid: false,
	}
}

func nullableText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{
			Valid: false,
		}
	}

	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}
