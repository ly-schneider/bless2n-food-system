package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/internal/config"

	"go.uber.org/zap"
)

type ElvantoService interface {
	SearchPeople(ctx context.Context) ([]ElvantoPerson, error)
	IsConfigured() bool
}

type ElvantoPerson struct {
	ID        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type elvantoService struct {
	cfg    config.ElvantoConfig
	client *http.Client
	logger *zap.Logger
}

func NewElvantoService(cfg config.Config, logger *zap.Logger) ElvantoService {
	return &elvantoService{
		cfg: cfg.Elvanto,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (s *elvantoService) IsConfigured() bool {
	return s.cfg.APIKey != ""
}

func (s *elvantoService) SearchPeople(ctx context.Context) ([]ElvantoPerson, error) {
	if !s.IsConfigured() {
		s.logger.Warn("elvanto not configured")
		return nil, fmt.Errorf("elvanto not configured")
	}

	s.logger.Info("searching elvanto people", zap.String("groupId", s.cfg.GroupID))

	reqBody := map[string]any{
		"page":      1,
		"page_size": 100,
		"search": map[string]any{
			"groups": s.cfg.GroupID,
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.elvanto.com/v1/people/search.json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(s.cfg.APIKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("elvanto request failed", zap.Error(err))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("elvanto API error", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return nil, fmt.Errorf("elvanto API returned %d", resp.StatusCode)
	}

	var result elvantoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("elvanto decode error", zap.Error(err))
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Status != "ok" {
		s.logger.Error("elvanto status not ok", zap.String("status", result.Status))
		return nil, fmt.Errorf("elvanto error: %s", result.Status)
	}

	s.logger.Info("elvanto returned people", zap.Int("count", len(result.People.Person)))
	return result.People.Person, nil
}

type elvantoResponse struct {
	Status string `json:"status"`
	People struct {
		Person []ElvantoPerson `json:"person"`
	} `json:"people"`
}
