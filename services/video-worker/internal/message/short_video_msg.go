package message

type ShortVideoFetchMessage struct {
	JobID              string `json:"jobId"`
	Provider           string `json:"provider"`
	Query              string `json:"query"`
	Orientation        string `json:"orientation"`
	Limit              int    `json:"limit"`
	Page               int    `json:"page"`
	MinDurationSeconds int    `json:"minDurationSeconds"`
	MaxDurationSeconds int    `json:"maxDurationSeconds"`
}
