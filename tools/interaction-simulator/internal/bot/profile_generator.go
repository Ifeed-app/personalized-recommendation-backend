package bot

import (
	"math/rand"

	"github.com/google/uuid"
)

var allCategories = []string{
	"fitness",
	"food",
	"travel",
	"nature",
	"cars",
	"technology",
	"animals",
	"sports",
	"fashion",
	"music",
	"architecture",
	"ocean",
}

func GenerateProfile() Profile {
	categories := append([]string(nil), allCategories...)

	rand.Shuffle(len(categories), func(i, j int) {
		categories[i], categories[j] = categories[j], categories[i]
	})

	personality := randomPersonality()

	switch personality {

	case PersonalityFocused:
		return Profile{
			UserID:      uuid.NewString(),
			Personality: personality,
			PreferredCategories: []string{
				categories[0],
				categories[1],
			},
			DislikedCategories: []string{
				categories[2],
				categories[3],
			},
		}

	case PersonalityBroad:
		return Profile{
			UserID:      uuid.NewString(),
			Personality: personality,
			PreferredCategories: []string{
				categories[0],
				categories[1],
				categories[2],
				categories[3],
			},
			DislikedCategories: []string{
				categories[4],
			},
		}

	case PersonalityPicky:
		return Profile{
			UserID:      uuid.NewString(),
			Personality: personality,
			PreferredCategories: []string{
				categories[0],
			},
			DislikedCategories: []string{
				categories[1],
				categories[2],
				categories[3],
			},
		}

	case PersonalityExplorer:
		return Profile{
			UserID:      uuid.NewString(),
			Personality: personality,
			PreferredCategories: []string{
				categories[0],
				categories[1],
				categories[2],
			},
			DislikedCategories: []string{},
		}
	}

	panic("unknown personality")
}

func randomPersonality() Personality {
	switch rand.Intn(4) {
	case 0:
		return PersonalityFocused
	case 1:
		return PersonalityBroad
	case 2:
		return PersonalityPicky
	default:
		return PersonalityExplorer
	}
}
