package upload_grpc

import (
	uploadv1 "backend-diplomski/proto/gen/upload/v1"
	"backend-diplomski/services/upload-service/db/sqlc"
	"backend-diplomski/services/upload-service/internal/storage"
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UploadServer struct {
	uploadv1.UnimplementedUploadServiceServer
	storage *storage.MinioStorage
	queries *sqlc.Queries
}

func NewUploadServer(storage *storage.MinioStorage, queries *sqlc.Queries) *UploadServer {
	return &UploadServer{
		storage: storage,
		queries: queries,
	}
}

func (s *UploadServer) UploadVideo(stream uploadv1.UploadService_UploadVideoServer) error {
	firstReq, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive metadata: %w", err)
	}
	metadata := firstReq.GetMetadata()
	if metadata == nil {
		return fmt.Errorf("first message must containt metadata")
	}
	source := sanitizePathPart(metadata.Source)
	externalID := sanitizePathPart(metadata.ExternalId)
	fileName := strings.TrimSpace(metadata.FileName)
	contentType := strings.TrimSpace(metadata.ContentType)
	objectKey := strings.TrimSpace(metadata.ObjectKey)

	if source == "" {
		source = "unknown"
	}
	if externalID == "" {
		return fmt.Errorf("external id is req")
	}
	if fileName == "" {
		fileName = externalID + ".mp4"
	}
	if contentType == "" {
		contentType = "video/mp4"
	}
	if objectKey == "" {
		extensions := filepath.Ext(fileName)
		if extensions == "" {
			extensions = ".mp4"
		}
		objectKey = fmt.Sprintf(
			"videos/imported/%s/%s/original%s",
			source,
			externalID,
			extensions,
		)
	}
	var buffer bytes.Buffer
	var sizeBytes int64

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %w", err)
		}
		chunk := req.GetChunk()
		if len(chunk) == 0 {
			continue
		}
		n, err := buffer.Write(chunk)
		if err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		sizeBytes += int64(n)
	}
	if sizeBytes == 0 {
		return fmt.Errorf("uploaded video is empty")
	}
	uploadID := uuid.New()
	uploadRecord, err := s.queries.CreateUpload(stream.Context(), sqlc.CreateUploadParams{
		ID:         toPgUUID(uploadID),
		UserID:     nullPgUUID(),
		UploadType: "VIDEO",
		Status:     "PENDING",
	})
	if err != nil {
		return fmt.Errorf("failed to create upload record: %w", err)
	}
	err = s.storage.UploadObject(stream.Context(), objectKey, bytes.NewReader(buffer.Bytes()), sizeBytes, contentType)
	if err != nil {
		_, _ = s.queries.UpdateUploadStatus(stream.Context(), sqlc.UpdateUploadStatusParams{
			ID:     uploadRecord.ID,
			Status: "FAILED",
		})
		return fmt.Errorf("failed to upload object to MinIO: %w", err)
	}

	log.Printf("uploaded video to MinIO bucket=%s objectKey=%s sizeBytes=%d contentType=%s",
		s.storage.Bucket(),
		objectKey,
		sizeBytes,
		contentType,
	)
	assetID := uuid.New()
	assetRecord, err := s.queries.CreateMediaAsset(stream.Context(), sqlc.CreateMediaAssetParams{
		ID:         toPgUUID(assetID),
		UploadID:   uploadRecord.ID,
		OwnerID:    nullPgUUID(),
		AssetType:  "VIDEO",
		StorageKey: objectKey,
		FileName:   fileName,
		MimeType:   contentType,
		SizeBytes:  sizeBytes,
		Checksum:   nullPgText(),
		Status:     "COMPLETED",
	})
	if err != nil {
		_, _ = s.queries.UpdateUploadStatus(stream.Context(), sqlc.UpdateUploadStatusParams{
			ID:     uploadRecord.ID,
			Status: "FAILED",
		})
		return fmt.Errorf("failed to create media asset record: %w", err)
	}
	_, err = s.queries.UpdateUploadStatus(stream.Context(), sqlc.UpdateUploadStatusParams{
		ID:     uploadRecord.ID,
		Status: "COMPLETED",
	})
	if err != nil {
		return fmt.Errorf("failed to update upload status: %w", err)
	}
	return stream.SendAndClose(&uploadv1.UploadVideoResponse{
		Bucket:           s.storage.Bucket(),
		ObjectKey:        objectKey,
		MimeType:         contentType,
		SizeBytes:        sizeBytes,
		OriginalFileName: fileName,
		UploadId:         uuid.UUID(uploadRecord.ID.Bytes).String(),
		AssetId:          uuid.UUID(assetRecord.ID.Bytes).String(),
	})
}

func sanitizePathPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "..", "")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
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

func nullPgText() pgtype.Text {
	return pgtype.Text{
		Valid: false,
	}
}
