package message

type ShortVideoFetchMessage struct {
	JobID       string `json:"jobId"`
	Source      string `json:"source"`
	VideoURL    string `json:"videoUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
