package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ahhcash/vsearch/config"
	"github.com/ahhcash/vsearch/models"
	"github.com/labstack/echo/v4"
)

type SearchHandler struct {
	cfg *config.Config
}

func NewSearchHandler(cfg *config.Config) *SearchHandler {
	return &SearchHandler{cfg: cfg}
}

func (h *SearchHandler) Search(c echo.Context) error {
	var req models.SearchReq
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

	var searchResp models.SearchResp
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse response")
	}

	return c.JSON(http.StatusOK, searchResp)
}

func validateSearchRequest(req *models.SearchReq) error {
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
		case "text", "url":
		default:
			return fmt.Errorf("query %d: invalid type %s", i, query.Type)
		}
	}

	return nil
}
