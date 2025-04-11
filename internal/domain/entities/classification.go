package entities

type ClassificationResult struct {
	Priority       string           `json:"priority"`
	Category       string           `json:"category"`
	Labels         []string         `json:"labels"`
	SuggestedTasks []*SuggestedTask `json:"suggestedTasks"`
}

type SuggestedTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}
