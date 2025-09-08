package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RepositoryInfo struct {
	Name     string    `json:"name"`
	FullName string    `json:"full_name,omitempty"`
	Tags     []TagInfo `json:"tags,omitempty"`
}

type TagInfo struct {
	Name   string `json:"name"`
	Digest string `json:"digest,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

type ListOptions struct {
	Page        int  `json:"page,omitempty"`
	PageSize    int  `json:"page_size,omitempty"`
	IncludeTags bool `json:"include_tags,omitempty"`
}

func DefaultListOptions() *ListOptions {
	return &ListOptions{Page: 1, PageSize: 50}
}

type RepositoryLister struct {
	registry *GiteaRegistry
}

func NewRepositoryLister(registry *GiteaRegistry) *RepositoryLister {
	return &RepositoryLister{registry}
}

func (l *RepositoryLister) ListRepositories(ctx context.Context, opts *ListOptions) ([]RepositoryInfo, error) {
	if opts == nil { opts = DefaultListOptions() }

	catalogURL := fmt.Sprintf("%s/v2/_catalog", l.registry.baseURL)
	if opts.PageSize > 0 { catalogURL += "?n=" + strconv.Itoa(opts.PageSize) }

	req, err := http.NewRequestWithContext(ctx, "GET", catalogURL, nil)
	if err != nil { return nil, err }
	if authHeader, err := l.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := l.registry.httpClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode) }

	var catalogResponse struct { Repositories []string `json:"repositories"` }
	if err := json.NewDecoder(resp.Body).Decode(&catalogResponse); err != nil { return nil, err }

	repositories := make([]RepositoryInfo, len(catalogResponse.Repositories))
	for i, repoName := range catalogResponse.Repositories {
		repo := RepositoryInfo{Name: repoName, FullName: repoName}
		if opts.IncludeTags {
			if tags, err := l.ListTags(ctx, repoName, opts); err == nil { repo.Tags = tags }
		}
		repositories[i] = repo
	}
	return repositories, nil
}

func (l *RepositoryLister) ListTags(ctx context.Context, repository string, opts *ListOptions) ([]TagInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v2/%s/tags/list", l.registry.baseURL, repository), nil)
	if err != nil { return nil, err }
	if authHeader, err := l.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := l.registry.httpClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound { return []TagInfo{}, nil }
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode) }

	var tagsResponse struct { Tags []string `json:"tags"` }
	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil { return nil, err }

	tags := make([]TagInfo, len(tagsResponse.Tags))
	for i, tagName := range tagsResponse.Tags { tags[i] = TagInfo{Name: tagName} }
	return tags, nil
}

func (l *RepositoryLister) SearchRepositories(ctx context.Context, query string, opts *ListOptions) ([]RepositoryInfo, error) {
	repos, err := l.ListRepositories(ctx, opts)
	if err != nil { return nil, err }
	
	var filtered []RepositoryInfo
	for _, repo := range repos {
		if strings.Contains(repo.Name, query) { filtered = append(filtered, repo) }
	}
	return filtered, nil
}