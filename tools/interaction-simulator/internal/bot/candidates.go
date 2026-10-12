package bot

import (
	"math/rand"

	contentv1 "backend-diplomski/proto/gen/content/v1"
)

func SelectRandomCandidates(
	candidates []*contentv1.RecommendationCandidate,
	count int,
) []*contentv1.RecommendationCandidate {

	if len(candidates) == 0 || count <= 0 {
		return nil
	}

	shuffled := append(
		[]*contentv1.RecommendationCandidate(nil),
		candidates...,
	)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	if count >= len(shuffled) {
		return shuffled
	}

	return shuffled[:count]
}
