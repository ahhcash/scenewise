package models

import "time"

type GrotleSearchRequest struct {
	Queries        []Query  `json:"queries"`
	Collections    []string `json:"collections"`
	Page           int      `json:"page"`
	OffsetPosition int      `json:"offset_position"`
}

type Query struct {
	Type           string `json:"type"` // "text", "base64"
	Value          string `json:"value"`
	EmbeddingModel string `json:"embedding_model"`
}

type GrotleSearchResult struct {
	ID           string  `json:"id"`
	URL          string  `json:"url"`
	ThumbnailURL string  `json:"thumbnailUrl,omitempty"`
	Title        string  `json:"title,omitempty"`
	Duration     float64 `json:"duration,omitempty"`

	StartTime float64 `json:"startTime,omitempty"`
	EndTime   float64 `json:"endTime,omitempty"`

	Score     float64 `json:"score"`
	MatchType string  `json:"matchType,omitempty"`

	Description string `json:"description,omitempty"`
	Transcript  string `json:"transcript,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	OriginalMetadata map[string]interface{} `json:"originalMetadata,omitempty"`
}

type GrotleSearchResponse struct {
	Results    []GrotleSearchResult `json:"results"`
	Pagination struct {
		CurrentPage  int  `json:"currentPage"`
		TotalPages   int  `json:"totalPages"`
		TotalResults int  `json:"totalResults"`
		HasMore      bool `json:"hasMore"`
	} `json:"pagination"`
	Debug map[string]interface{} `json:"debug,omitempty"`
}
