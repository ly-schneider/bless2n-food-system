package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/config"

	"go.uber.org/zap"
)

type AndroidRelease struct {
	VersionName string `json:"versionName"`
	VersionCode int    `json:"versionCode"`
	DownloadURL string `json:"downloadUrl"`
	SHA256      string `json:"sha256,omitempty"`
}

type AndroidUpdateService interface {
	GetLatestRelease(ctx context.Context) (*AndroidRelease, error)
}

type androidUpdateService struct {
	repoOwner string
	repoName  string
	logger    *zap.Logger

	mu        sync.Mutex
	cached    *AndroidRelease
	cachedAt  time.Time
	cacheTTL  time.Duration
}

func NewAndroidUpdateService(cfg config.AndroidConfig, logger *zap.Logger) AndroidUpdateService {
	parts := strings.SplitN(cfg.GitHubRepo, "/", 2)
	owner, repo := parts[0], parts[1]
	return &androidUpdateService{
		repoOwner: owner,
		repoName:  repo,
		logger:    logger,
		cacheTTL:  5 * time.Minute,
	}
}

func (s *androidUpdateService) GetLatestRelease(ctx context.Context) (*AndroidRelease, error) {
	s.mu.Lock()
	if s.cached != nil && time.Since(s.cachedAt) < s.cacheTTL {
		result := s.cached
		s.mu.Unlock()
		return result, nil
	}
	s.mu.Unlock()

	release, err := s.fetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cached = release
	s.cachedAt = time.Now()
	s.mu.Unlock()

	return release, nil
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (s *androidUpdateService) fetchLatestRelease(ctx context.Context) (*AndroidRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", s.repoOwner, s.repoName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("android update: create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("android update: fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("android update: GitHub API returned %d", resp.StatusCode)
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("android update: decode releases: %w", err)
	}

	for _, rel := range releases {
		if !strings.HasPrefix(rel.TagName, "android-production-v") {
			continue
		}

		versionStr := strings.TrimPrefix(rel.TagName, "android-production-v")
		versionCode, err := parseVersionCode(versionStr)
		if err != nil {
			s.logger.Warn("android update: invalid version in tag", zap.String("tag", rel.TagName), zap.Error(err))
			continue
		}

		var downloadURL string
		var sha256URL string
		for _, asset := range rel.Assets {
			if strings.HasSuffix(asset.Name, ".apk") {
				downloadURL = asset.BrowserDownloadURL
			}
			if strings.HasSuffix(asset.Name, ".sha256") {
				sha256URL = asset.BrowserDownloadURL
			}
		}

		if downloadURL == "" {
			continue
		}

		sha256 := ""
		if sha256URL != "" {
			sha256 = s.fetchSHA256(ctx, sha256URL)
		}

		return &AndroidRelease{
			VersionName: versionStr,
			VersionCode: versionCode,
			DownloadURL: downloadURL,
			SHA256:      sha256,
		}, nil
	}

	return nil, nil
}

func (s *androidUpdateService) fetchSHA256(ctx context.Context, url string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return ""
	}

	// sha256sum format: "<hash>  <filename>" or just "<hash>"
	parts := strings.Fields(strings.TrimSpace(string(body)))
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func parseVersionCode(version string) (int, error) {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	pat, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	return maj*10000 + min*100 + pat, nil
}
