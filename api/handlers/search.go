package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ahhcash/vsearch/models"
	"github.com/ahhcash/vsearch/models/mixpeek"
	"io"
	"net/http"

	"github.com/ahhcash/vsearch/config"
	"github.com/labstack/echo/v4"
)

type SearchHandler struct {
	cfg *config.Config
}

func NewSearchHandler(cfg *config.Config) *SearchHandler {
	return &SearchHandler{cfg: cfg}
}

func (h *SearchHandler) Search(c echo.Context) error {
	var req mixpeek.MixpeekSearchReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := validateSearchRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if len(req.Collections) == 0 {
		req.Collections = []string{h.cfg.CollectionName}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	mixpeekReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/features/search", h.cfg.MixpeekBaseURL),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request")
	}

	mixpeekReq.Header.Set("Content-Type", "application/json")
	mixpeekReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.cfg.MixpeekAPIKey))

	client := &http.Client{}
	resp, err := client.Do(mixpeekReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to perform search")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(resp.StatusCode, fmt.Sprintf("Mixpeek API error: %s", string(respBody)))
	}

	var mixpeekResp mixpeek.MixpeekSearchResp
	if err := json.Unmarshal(respBody, &mixpeekResp); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse response")
	}

	videoResults := make([]models.VideoSearchResult, 0)
	for _, result := range mixpeekResp.Results {
		url := extractURL(result)

		videoResult := models.VideoSearchResult{
			ID:          result.FeatureID,
			URL:         url,
			Score:       result.Score,
			StartTime:   getFloatValue(result.StartTime),
			EndTime:     getFloatValue(result.EndTime),
			Description: result.Description,
			Transcript:  result.Transcription,
			CreatedAt:   result.CreatedAt,

			MatchType: determineMatchType(result),

			OriginalMetadata: result.Metadata,
		}

		fileData := result.FileData
		if duration, ok := fileData["duration"].(float64); ok {
			videoResult.Duration = duration
		}
		if title, ok := fileData["file_name"].(string); ok {
			videoResult.Title = title
		}
		if thumbnail, ok := fileData["thumbnail"].(string); ok {
			videoResult.ThumbnailURL = thumbnail
		}

		videoResults = append(videoResults, videoResult)
	}

	response := models.VideoSearchResponse{
		Results: videoResults,
		Pagination: struct {
			CurrentPage  int  `json:"currentPage"`
			TotalPages   int  `json:"totalPages"`
			TotalResults int  `json:"totalResults"`
			HasMore      bool `json:"hasMore"`
		}{
			CurrentPage:  1, // Extract from request/response
			TotalPages:   parseTotalPages(mixpeekResp),
			TotalResults: mixpeekResp.Total,
			HasMore:      hasMoreResults(mixpeekResp),
		},
	}

	return c.JSON(http.StatusOK, response)
}

func getFloatValue(ptr *float64) float64 {
	if ptr == nil {
		return 0.0
	}

	return *ptr
}

func validateSearchRequest(req *mixpeek.MixpeekSearchReq) error {
	if len(req.Queries) == 0 {
		return fmt.Errorf("at least one query is required")
	}

	for i, query := range req.Queries {
		if query.Value == "" {
			return fmt.Errorf("query %d: value is required", i)
		}
		if query.Type == "" {
			return fmt.Errorf("query %d: type is required", i)
		}
		switch query.Type {
		case "text", "url", "base64":
		default:
			return fmt.Errorf("query %d: invalid type %s", i, query.Type)
		}
	}

	return nil
}

func extractURL(result mixpeek.SearchResult) string {
	if url, ok := result.FileData["url"].(string); ok {
		return url
	}

	if url, ok := result.Metadata["url"].(string); ok {
		return url
	}

	return ""
}

func determineMatchType(result mixpeek.SearchResult) string {
	return "visual" // Default for now
}

func parseTotalPages(resp mixpeek.MixpeekSearchResp) int {
	if resp.Pagination != nil {
		if totalPages, ok := resp.Pagination["total_pages"].(float64); ok {
			return int(totalPages)
		}

		if total, ok := resp.Pagination["total"].(float64); ok {
			pageSize := 10
			if size, ok := resp.Pagination["page_size"].(float64); ok {
				pageSize = int(size)
			}
			return int((total + float64(pageSize) - 1) / float64(pageSize))
		}
	}

	return 0
}

func hasMoreResults(resp mixpeek.MixpeekSearchResp) bool {
	if resp.Pagination != nil {
		if nextPage, ok := resp.Pagination["next_page"].(string); ok {
			return nextPage != ""
		}

		if currentPage, ok := resp.Pagination["page"].(float64); ok {
			totalPages := parseTotalPages(resp)
			return int(currentPage) < totalPages
		}
	}
	return false
}
