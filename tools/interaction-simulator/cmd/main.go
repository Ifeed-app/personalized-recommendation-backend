package main

import (
	"context"
	"log"
	"os"
	"time"

	"backend-diplomski/tools/interaction-simulator/internal/bot"
	"backend-diplomski/tools/interaction-simulator/internal/client"
)

const (
	botCount           = 109
	videosPerBot       = 200
	candidatePoolLimit = 2000
)

func main() {
	interactionAddress := os.Getenv("INTERACTION_SERVICE_GRPC_URL")
	contentAddress := os.Getenv("CONTENT_SERVICE_GRPC_URL")
	serviceToken := os.Getenv("INTERACTION_SIMULATOR_TOKEN")

	if interactionAddress == "" {
		interactionAddress = "localhost:9070"
	}

	if contentAddress == "" {
		contentAddress = "localhost:9080"
	}

	if serviceToken == "" {
		log.Fatal("INTERACTION_SIMULATOR_TOKEN is required")
	}

	interactionClient, err := client.NewInteractionClient(
		interactionAddress,
		serviceToken,
	)
	if err != nil {
		log.Fatalf("failed to create interaction client: %v", err)
	}
	defer interactionClient.Close()

	contentClient, err := client.NewContentClient(
		contentAddress,
		serviceToken,
	)
	if err != nil {
		log.Fatalf("failed to create content client: %v", err)
	}
	defer contentClient.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Minute,
	)
	defer cancel()

	candidates, err := contentClient.ListRecommendationCandidates(
		ctx,
		candidatePoolLimit,
	)
	if err != nil {
		log.Fatalf("failed to load candidates: %v", err)
	}

	log.Printf(
		"loaded %d candidate videos",
		len(candidates),
	)

	profiles := bot.GeneratePopulation(botCount)

	simulator := bot.NewSimulator(
		interactionClient,
	)

	for index, profile := range profiles {
		log.Printf(
			"starting bot %d/%d user=%s personality=%s preferred=%v disliked=%v",
			index+1,
			len(profiles),
			profile.UserID,
			profile.Personality,
			profile.PreferredCategories,
			profile.DislikedCategories,
		)

		selectedVideos := bot.SelectRandomCandidates(
			candidates,
			videosPerBot,
		)

		for _, video := range selectedVideos {
			if err := simulator.SimulateVideo(
				ctx,
				profile,
				video,
			); err != nil {
				log.Printf(
					"simulation error bot=%s video=%s: %v",
					profile.UserID,
					video.GetVideoId(),
					err,
				)
				continue
			}

			time.Sleep(50 * time.Millisecond)
		}

		log.Printf(
			"finished bot %d/%d user=%s",
			index+1,
			len(profiles),
			profile.UserID,
		)
	}

	log.Println("population simulation finished")
}
