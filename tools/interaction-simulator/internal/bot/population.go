package bot

func GeneratePopulation(count int) []Profile {
	profiles := make([]Profile, 0, count)

	for i := 0; i < count; i++ {
		profiles = append(
			profiles,
			GenerateProfile(),
		)
	}

	return profiles
}
