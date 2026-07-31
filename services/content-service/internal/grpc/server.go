package grpc

import (
	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/services/content-service/db/sqlc"
	"backend-diplomski/services/content-service/internal/storage"
	"context"
	"errors"
	"fmt"
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

func (s *Server) CreateVideo(ctx context.Context, req *contentv1.CreateVideoRequest) (*contentv1.CreateVideoResponse, error) {
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
	existingVideo, err := s.queries.GetVideoByStorageKey(ctx, req.StroageKey)
	if err == nil {
		return &contentv1.CreateVideoResponse{
			Id:         uuid.UUID(existingVideo.ID.Bytes).String(),
			Title:      existingVideo.Title,
			StroageKey: existingVideo.StorageKey,
			Status:     existingVideo.Status,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing video: %w", err)
	}
	videoID := uuid.New()
	video, err := s.queries.CreateVideo(ctx, sqlc.CreateVideoParams{
		ID:              toPgUUID(videoID),
		CreatorID:       nullPgUUID(),
		Title:           req.Title,
		Description:     nullableText(req.Description),
		StorageKey:      req.StroageKey,
		ThumbnailKey:    nullableText(req.ThumbnailKey),
		MimeType:        req.MimeType,
		DurationSeconds: req.DurationSecond,
		Status:          req.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
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
