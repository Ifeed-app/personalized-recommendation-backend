package recommendation

type InteractionScoreInput struct {
	Liked               bool
	Unliked             bool
	Skipped             bool
	CompletedCount      int32
	TotalWatchedSeconds int32
	DurationSeconds     int32
}

func CalculateInteractionScore(
	input InteractionScoreInput,
) float64 {
	score := 0.0

	if input.Liked {
		score += 3.0
	}

	if input.Unliked {
		score -= 3.0
	}

	if input.Skipped {
		score -= 1.5
	}

	if input.DurationSeconds > 0 {
		watchRatio :=
			float64(input.TotalWatchedSeconds) /
				float64(input.DurationSeconds)

		switch {
		case watchRatio >= 0.8:
			score += 2.0

		case watchRatio >= 0.5:
			score += 1.0

		case watchRatio < 0.2:
			score -= 1.0
		}
	}

	return score
}
func CalculateCandidateScore(
	categoryScores map[string]float64,
	categories []string,
) float64 {
	score := 0.0

	for _, category := range categories {
		score += categoryScores[category]
	}

	return score
}
