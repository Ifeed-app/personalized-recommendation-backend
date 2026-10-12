package bot

type Personality string

const (
	PersonalityFocused  Personality = "focused"
	PersonalityBroad    Personality = "broad"
	PersonalityPicky    Personality = "picky"
	PersonalityExplorer Personality = "explorer"
)

type Profile struct {
	UserID              string
	Personality         Personality
	PreferredCategories []string
	DislikedCategories  []string
}
