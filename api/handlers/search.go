package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ahhcash/vsearch/models"
	"github.com/ahhcash/vsearch/models/mixpeek"
	"io"
	"net/http"
	"net/url"

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
	var grotleReq models.GrotleSearchRequest
	if err := c.Bind(&grotleReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := validateSearchRequest(&grotleReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mixpeekQueries := make([]mixpeek.Query, 0)
	for _, q := range grotleReq.Queries {

		mixpeekQueries = append(mixpeekQueries, mixpeek.Query{
			Type:           q.Type,
			Value:          q.Value,
			EmbeddingModel: "multimodal",
		})

		if "text" == q.Type {
			mixpeekQueries = append(mixpeekQueries, mixpeek.Query{
				Type:           q.Type,
				Value:          q.Value,
				EmbeddingModel: "text",
			})
		}
	}

	if nil == grotleReq.Collections || len(grotleReq.Collections) == 0 {
		grotleReq.Collections = []string{h.cfg.CollectionName}
	}

	mixpeekReq := mixpeek.MixpeekSearchReq{
		Queries:     mixpeekQueries,
		Collections: grotleReq.Collections,
	}

	baseURL := fmt.Sprintf("%s/features/search", h.cfg.MixpeekBaseURL)
	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to construct API URL")
	}

	q := apiURL.Query()
	q.Set("page", fmt.Sprintf("%d", grotleReq.Page))
	q.Set("offset_position", fmt.Sprintf("%d", grotleReq.OffsetPosition))
	apiURL.RawQuery = q.Encode()
	body, err := json.Marshal(mixpeekReq)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	// Create request to Mixpeek API
	req, err := http.NewRequest(
		http.MethodPost,
		apiURL.String(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.cfg.MixpeekAPIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
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

	videoResults := make([]models.GrotleSearchResult, 0)
	for _, result := range mixpeekResp.Results {
		urlExtracted := extractURL(result)

		videoResult := models.GrotleSearchResult{
			ID:               result.FeatureID,
			URL:              urlExtracted,
			Score:            result.Score,
			StartTime:        getFloatValue(result.StartTime),
			EndTime:          getFloatValue(result.EndTime),
			Description:      result.Description,
			Transcript:       result.Transcription,
			CreatedAt:        result.CreatedAt,
			MatchType:        determineMatchType(result),
			OriginalMetadata: result.Metadata,
		}

		if fileData := result.FileData; fileData != nil {
			if duration, ok := fileData["duration"].(float64); ok {
				videoResult.Duration = duration
			}
			if title, ok := fileData["file_name"].(string); ok {
				videoResult.Title = title
			}
			if thumbnail, ok := fileData["thumbnail"].(string); ok {
				videoResult.ThumbnailURL = thumbnail
			}
		}

		videoResults = append(videoResults, videoResult)
	}

	response := models.GrotleSearchResponse{
		Results: videoResults,
		Pagination: struct {
			CurrentPage  int  `json:"currentPage"`
			TotalPages   int  `json:"totalPages"`
			TotalResults int  `json:"totalResults"`
			HasMore      bool `json:"hasMore"`
		}{
			CurrentPage:  getCurrentPage(mixpeekResp),
			TotalPages:   parseTotalPages(mixpeekResp),
			TotalResults: getTotalResults(mixpeekResp),
			HasMore:      hasMoreResults(mixpeekResp),
		},
	}

	return c.JSON(http.StatusOK, response)
}

func getTotalResults(resp mixpeek.MixpeekSearchResp) int {
	if resp.Pagination != nil {
		if total, ok := resp.Pagination["total"].(float64); ok {
			return int(total)
		}
	}
	return 0
}

// Also add this helper function to get current page
func getCurrentPage(resp mixpeek.MixpeekSearchResp) int {
	if resp.Pagination != nil {
		if page, ok := resp.Pagination["page"].(float64); ok {
			return int(page)
		}
		if page, ok := resp.Pagination["current_page"].(float64); ok {
			return int(page)
		}
	}
	return 1
}

func getFloatValue(ptr *float64) float64 {
	if ptr == nil {
		return 0.0
	}

	return *ptr
}

func validateSearchRequest(req *models.GrotleSearchRequest) error {
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
	if uri, ok := result.FileData["url"].(string); ok {
		return uri
	}

	if uri, ok := result.Metadata["url"].(string); ok {
		return uri
	}

	return ""
}

func determineMatchType(result mixpeek.SearchResult) string {
	if result.OriginalValues != nil {
		_, hasMultimodal := result.OriginalValues["multimodal"]
		_, hasText := result.OriginalValues["text"]
		_, hasImage := result.OriginalValues["image"]
		_, hasVideo := result.OriginalValues["video"]

		if hasMultimodal {
			return "multimodal"
		} else if hasVideo {
			return "video"
		} else if hasImage {
			return "visual"
		} else if hasText {
			return "text"
		}
	}

	switch result.Modality {
	case "video":
		return "video"
	case "image":
		return "visual"
	case "text":
		return "text"
	default:
		return "unknown"
	}
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
