package main

type SearchReq struct {
	Queries          []Query        `json:"queries"`
	Collections      []string       `json:"collections"`
	Filters          map[string]any `json:"filters"`
	GroupBy          map[string]any `json:"group_by"`
	Sort             map[string]any `json:"sort"`
	Select           []string       `json:"select"`
	RerankingOptions map[string]any `json:"reranking_options"`
	ReturnURL        bool           `json:"return_url"`
	SessionID        map[string]any `json:"session_id"`
}

type Query struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	EmbeddingModel string `json:"embedding_model"`
}
