package mixpeek

import "time"

type MixpeekSearchReq struct {
	Queries          []Query        `json:"queries"`
	Collections      []string       `json:"collections"`
	Filters          map[string]any `json:"filters,omitempty"`
	GroupBy          map[string]any `json:"group_by,omitempty"`
	Sort             map[string]any `json:"sort,omitempty"`
	Select           []string       `json:"select,omitempty"`
	RerankingOptions map[string]any `json:"reranking_options,omitempty"`
	ReturnURL        bool           `json:"return_url"`
	SessionID        map[string]any `json:"session_id"`
}

type Query struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	EmbeddingModel string `json:"embedding_model"`
}

type MixpeekSearchResp struct {
	Results    []SearchResult `json:"results"`
	Pagination map[string]any `json:"pagination"`
	Total      int            `json:"total"`
}

type SearchResult struct {
	FeatureID    string    `json:"feature_id"`
	AssetID      string    `json:"asset_id"`
	CollectionID string    `json:"collection_id"`
	Score        float64   `json:"score"`
	CreatedAt    time.Time `json:"created_at"`

	FileData       map[string]any `json:"file_data"`
	Metadata       map[string]any `json:"metadata"`
	OriginalValues map[string]any `json:"original_values"`

	StartTime     *float64 `json:"start_time,omitempty"`
	EndTime       *float64 `json:"end_time,omitempty"`
	IntervalSec   *int     `json:"interval_sec,omitempty"`
	Modality      string   `json:"modality,omitempty"`
	Type          string   `json:"type,omitempty"`
	Transcription string   `json:"transcription,omitempty"`
	Description   string   `json:"description,omitempty"`
}
