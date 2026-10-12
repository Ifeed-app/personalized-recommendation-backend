package bot

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/tools/interaction-simulator/internal/client"
)

type Simulator struct {
	interactionClient *client.InteractionClient
}

func NewSimulator(
	interactionClient *client.InteractionClient,
) *Simulator {
	return &Simulator{
		interactionClient: interactionClient,
	}
}

func (s *Simulator) SimulateVideo(
	ctx context.Context,
	profile Profile,
	video *contentv1.RecommendationCandidate,
) error {

	categories := make([]string, 0, len(video.GetCategories()))

	for _, category := range video.GetCategories() {
		categories = append(categories, category.GetName())
	}

	decision := Decide(profile, categories)

	duration := video.GetDurationSecond()
	if duration <= 0 {
		return nil
	}

	watchedSeconds := int32(
		math.Round(float64(duration) * decision.WatchRatio),
	)

	if watchedSeconds < 1 {
		watchedSeconds = 1
	}

	if watchedSeconds > duration {
		watchedSeconds = duration
	}

	completed := decision.WatchRatio >= 0.90

	startedAt := time.Now().UTC()
	endedAt := startedAt.Add(
		time.Duration(watchedSeconds) * time.Second,
	)

	err := s.interactionClient.RecordWatchSession(
		ctx,
		profile.UserID,
		video.GetVideoId(),
		watchedSeconds,
		completed,
		startedAt.Format(time.RFC3339),
		endedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf(
			"failed watch session video=%s: %w",
			video.GetVideoId(),
			err,
		)
	}

	if decision.HasInteraction {
		err = s.interactionClient.CreateInteraction(
			ctx,
			profile.UserID,
			video.GetVideoId(),
			decision.InteractionType,
		)
		if err != nil {
			return fmt.Errorf(
				"failed interaction video=%s: %w",
				video.GetVideoId(),
				err,
			)
		}
	}

	log.Printf(
		"bot=%s video=%s categories=%v duration=%d watched=%d ratio=%.2f interaction=%s",
		profile.UserID,
		video.GetVideoId(),
		categories,
		duration,
		watchedSeconds,
		decision.WatchRatio,
		decision.InteractionType.String(),
	)

	return nil
}
