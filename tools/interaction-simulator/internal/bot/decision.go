package bot

import (
	"math/rand"

	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
)

type Decision struct {
	WatchRatio      float64
	InteractionType interactionv1.InteractionType
	HasInteraction  bool
}

func Decide(
	profile Profile,
	categories []string,
) Decision {

	preferred := containsAny(
		categories,
		profile.PreferredCategories,
	)

	disliked := containsAny(
		categories,
		profile.DislikedCategories,
	)

	switch profile.Personality {

	case PersonalityFocused:
		return decideFocused(preferred, disliked)

	case PersonalityBroad:
		return decideBroad(preferred, disliked)

	case PersonalityPicky:
		return decidePicky(preferred, disliked)

	case PersonalityExplorer:
		return decideExplorer(preferred)

	default:
		return decideBroad(preferred, disliked)
	}
}

func containsAny(
	videoCategories []string,
	targetCategories []string,
) bool {
	for _, videoCategory := range videoCategories {
		for _, targetCategory := range targetCategories {
			if videoCategory == targetCategory {
				return true
			}
		}
	}

	return false
}

func decideFocused(
	preferred bool,
	disliked bool,
) Decision {

	r := rand.Float64()

	if preferred {
		return Decision{
			WatchRatio:      0.75 + rand.Float64()*0.25,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_LIKE,
			HasInteraction:  r < 0.80,
		}
	}

	if disliked {
		return Decision{
			WatchRatio:      0.05 + rand.Float64()*0.20,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_SKIP,
			HasInteraction:  r < 0.80,
		}
	}

	return Decision{
		WatchRatio:      0.25 + rand.Float64()*0.40,
		InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_SKIP,
		HasInteraction:  r < 0.25,
	}
}

func decideBroad(
	preferred bool,
	disliked bool,
) Decision {

	r := rand.Float64()

	if preferred {
		return Decision{
			WatchRatio:      0.65 + rand.Float64()*0.35,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_LIKE,
			HasInteraction:  r < 0.60,
		}
	}

	if disliked {
		return Decision{
			WatchRatio:      0.10 + rand.Float64()*0.25,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_SKIP,
			HasInteraction:  r < 0.60,
		}
	}

	return Decision{
		WatchRatio: 0.35 + rand.Float64()*0.45,
	}
}

func decidePicky(
	preferred bool,
	disliked bool,
) Decision {

	r := rand.Float64()

	if preferred {
		return Decision{
			WatchRatio:      0.70 + rand.Float64()*0.30,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_LIKE,
			HasInteraction:  r < 0.45,
		}
	}

	if disliked {
		return Decision{
			WatchRatio:      0.03 + rand.Float64()*0.15,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_UNLIKE,
			HasInteraction:  r < 0.55,
		}
	}

	return Decision{
		WatchRatio:      0.15 + rand.Float64()*0.35,
		InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_SKIP,
		HasInteraction:  r < 0.50,
	}
}

func decideExplorer(
	preferred bool,
) Decision {

	r := rand.Float64()

	if preferred {
		return Decision{
			WatchRatio:      0.60 + rand.Float64()*0.40,
			InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_LIKE,
			HasInteraction:  r < 0.45,
		}
	}

	return Decision{
		WatchRatio:      0.40 + rand.Float64()*0.50,
		InteractionType: interactionv1.InteractionType_INTERACTION_TYPE_LIKE,
		HasInteraction:  r < 0.12,
	}
}
