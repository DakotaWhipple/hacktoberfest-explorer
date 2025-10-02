package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hacktober/internal/logger"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// Client wraps GitHub API client with additional functionality
type Client struct {
	client *github.Client
	ctx    context.Context
}

// Repository represents a GitHub repository with additional metadata
type Repository struct {
	*github.Repository
	RelevanceScore int
	Languages      []string
}

// Issue represents a GitHub issue with additional metadata
type Issue struct {
	*github.Issue
	Repository      *Repository
	DifficultyScore int
	RelevanceScore  int
}

// IssueStats contains statistics about issues in a repository
type IssueStats struct {
	Issues      []*Issue
	LabelCounts map[string]int
	TotalIssues int
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

// SearchHacktoberfestRepos searches for Hacktoberfest repositories with minimum stars
// SearchHacktoberfestRepos searches for Hacktoberfest repositories with minimum stars.
// It now returns both the collected repositories (limited by maxResults) and the
// total number of Hacktoberfest repositories matching the base criteria (without
// language filters) so the UI can show users how many exist in total.
func (c *Client) SearchHacktoberfestRepos(minStars int, languages []string, maxResults int) ([]*Repository, int, error) {
	return c.SearchHacktoberfestReposWithPage(minStars, languages, maxResults, 1)
}

// SearchHacktoberfestReposWithPage searches for Hacktoberfest repositories with pagination support.
func (c *Client) SearchHacktoberfestReposWithPage(minStars int, languages []string, maxResults int, page int) ([]*Repository, int, error) {
	start := time.Now()
	logger.Info(fmt.Sprintf("Starting repository search with languages: %v, page: %d", languages, page))

	var allRepos []*Repository
	repoMap := make(map[string]*Repository) // To deduplicate repos

	// If no languages specified, search without language filter
	if len(languages) == 0 {
		languages = []string{""}
	}

	// First, get a global total (without language filter) so user sees overall scale
	globalQuery := fmt.Sprintf("topic:hacktoberfest stars:>=%d archived:false", minStars)
	logger.Info(fmt.Sprintf("Getting global repository count with query: %s", globalQuery))
	globalOpts := &github.SearchOptions{Sort: "stars", Order: "desc", ListOptions: github.ListOptions{PerPage: 1}}
	globalResult, globalResp, globalErr := c.client.Search.Repositories(c.ctx, globalQuery, globalOpts)
	totalAvailable := 0
	if globalResp != nil {
		logger.LogAPIRequest("repositories/search_total", globalQuery, globalResp.StatusCode, time.Since(start))
		logger.Debug(fmt.Sprintf("(Total) Rate limit remaining: %d, resets at: %v", globalResp.Rate.Remaining, globalResp.Rate.Reset.Time))
	}
	if globalErr != nil {
		logger.ErrorWithErr("Failed to retrieve global total repository count", globalErr)
		// Continue with language searches even if global count fails
	} else if globalResult != nil && globalResult.Total != nil {
		totalAvailable = *globalResult.Total
		logger.Info(fmt.Sprintf("Global Hacktoberfest repositories total: %d", totalAvailable))
	} else {
		logger.Info("Global count query succeeded but no total available")
	}

	for _, lang := range languages {
		query := fmt.Sprintf("topic:hacktoberfest stars:>=%d archived:false sort:stars-desc", minStars)

		// Add language filter if specified
		if lang != "" {
			query += fmt.Sprintf(" language:%s", strings.ToLower(lang))
			logger.Debug(fmt.Sprintf("Searching for language: %s", lang))
		} else {
			logger.Debug("Searching without language filter")
		}

		logger.Info(fmt.Sprintf("Repository search query: %s", query))

		opts := &github.SearchOptions{
			Sort:  "stars",
			Order: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: min(100, maxResults), // GitHub API limit
			},
		}

		result, response, err := c.client.Search.Repositories(c.ctx, query, opts)

		if response != nil {
			logger.LogAPIRequest("repositories/search", query, response.StatusCode, time.Since(start))
			logger.Debug(fmt.Sprintf("Rate limit remaining: %d, resets at: %v",
				response.Rate.Remaining, response.Rate.Reset.Time))
		}

		if err != nil {
			logger.ErrorWithErr(fmt.Sprintf("Failed to search repositories for language: %s", lang), err)
			continue // Continue with other languages instead of failing completely
		}

		totalFound := 0
		if result.Total != nil {
			totalFound = *result.Total
		}

		logger.Info(fmt.Sprintf("Language %s search completed: %d total found, %d returned",
			lang, totalFound, len(result.Repositories)))

		// Process repositories for this language
		for _, repo := range result.Repositories {
			if repo.StargazersCount != nil && *repo.StargazersCount >= minStars {
				repoKey := fmt.Sprintf("%s/%s", *repo.Owner.Login, *repo.Name)

				// Skip archived repositories
				if repo.Archived != nil && *repo.Archived {
					logger.Debug(fmt.Sprintf("Repository %s is archived, skipping", repoKey))
					continue
				}

				// Skip if we already have this repo (from another language search)
				if _, exists := repoMap[repoKey]; exists {
					logger.Debug(fmt.Sprintf("Repository %s already found, skipping duplicate", repoKey))
					continue
				}

				r := &Repository{
					Repository: repo,
				}
				r.calculateRelevance(languages)
				repoMap[repoKey] = r

				logger.Debug(fmt.Sprintf("Repository processed: %s, stars: %d, archived: %v, relevance: %d",
					repoKey, *repo.StargazersCount, repo.Archived != nil && *repo.Archived, r.RelevanceScore))
			}
		}

		// Stop if we've reached our target
		if len(repoMap) >= maxResults {
			logger.Info(fmt.Sprintf("Reached maximum results (%d), stopping search", maxResults))
			break
		}
	}

	// Convert map to slice
	for _, repo := range repoMap {
		allRepos = append(allRepos, repo)
	}

	// Sort by relevance score (highest first)
	for i := 0; i < len(allRepos); i++ {
		for j := i + 1; j < len(allRepos); j++ {
			if allRepos[i].RelevanceScore < allRepos[j].RelevanceScore {
				allRepos[i], allRepos[j] = allRepos[j], allRepos[i]
			}
		}
	}

	// Limit to maxResults
	if len(allRepos) > maxResults {
		allRepos = allRepos[:maxResults]
	}

	duration := time.Since(start)
	logger.LogRepoSearch(fmt.Sprintf("languages: %v", languages), len(allRepos), len(allRepos), languages)
	logger.Info(fmt.Sprintf("Repository search completed: %d returned (limit %d), global total: %d, took %v", len(allRepos), maxResults, totalAvailable, duration))

	return allRepos, totalAvailable, nil
}

// GetRepositoryIssues fetches issues for a specific repository with label statistics
func (c *Client) GetRepositoryIssues(owner, repo string, labels []string, maxResults int) (*IssueStats, error) {
	start := time.Now()
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	logger.Info(fmt.Sprintf("Starting issue search for %s", repoName))

	opts := &github.IssueListByRepoOptions{
		State:     "open",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: min(maxResults, 100),
		},
	}

	logger.Debug(fmt.Sprintf("Making API call to list issues for %s with options: state=open, sort=updated, page=1, perPage=%d",
		repoName, opts.ListOptions.PerPage))

	issues, response, err := c.client.Issues.ListByRepo(c.ctx, owner, repo, opts)
	duration := time.Since(start)

	if response != nil {
		logger.LogAPIRequest("issues/list", repoName, response.StatusCode, duration)
		logger.Debug(fmt.Sprintf("Rate limit remaining: %d, resets at: %v",
			response.Rate.Remaining, response.Rate.Reset.Time))
		logger.Debug(fmt.Sprintf("Response headers - Link: %s, Last-Modified: %s",
			response.Header.Get("Link"), response.Header.Get("Last-Modified")))
	}

	if err != nil {
		logger.ErrorWithErr("Failed to fetch repository issues", err)
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	logger.Info(fmt.Sprintf("GitHub API returned %d items for %s (includes issues + PRs), took %v",
		len(issues), repoName, duration))

	result := make([]*Issue, 0, len(issues))
	labelCounts := make(map[string]int)
	prCount := 0

	logger.Debug(fmt.Sprintf("Processing %d items from GitHub API for %s", len(issues), repoName))

	for _, issue := range issues {
		logger.Debug(fmt.Sprintf("Processing item #%d: %s, has PR links: %v",
			*issue.Number, *issue.Title, issue.PullRequestLinks != nil))

		// Skip pull requests
		if issue.PullRequestLinks != nil {
			prCount++
			logger.Debug(fmt.Sprintf("Skipping pull request #%d: %s", *issue.Number, *issue.Title))
			continue
		}

		i := &Issue{
			Issue: issue,
		}
		i.calculateDifficulty()
		result = append(result, i)

		// Count all labels for statistics
		labelList := make([]string, 0, len(issue.Labels))
		for _, label := range issue.Labels {
			labelName := strings.ToLower(*label.Name)
			labelCounts[labelName]++
			labelList = append(labelList, labelName)
		}

		// Log each issue found with labels
		logger.Debug(fmt.Sprintf("Issue included #%d: %s, difficulty: %d, labels: [%s]",
			*issue.Number, *issue.Title, i.DifficultyScore, strings.Join(labelList, ", ")))
	}

	logger.Info(fmt.Sprintf("Processing complete for %s: %d total items, %d PRs skipped, %d actual issues, %d unique labels",
		repoName, len(issues), prCount, len(result), len(labelCounts)))

	stats := &IssueStats{
		Issues:      result,
		LabelCounts: labelCounts,
		TotalIssues: len(result),
	}

	logger.Info(fmt.Sprintf("Issue search completed for %s: returning %d issues with %d unique labels",
		repoName, len(result), len(labelCounts)))

	return stats, nil
}

// GetRepositoryLanguages fetches the languages used in a repository
func (c *Client) GetRepositoryLanguages(owner, repo string) ([]string, error) {
	languages, _, err := c.client.Repositories.ListLanguages(c.ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(languages))
	for lang := range languages {
		result = append(result, lang)
	}

	return result, nil
}

// calculateRelevance calculates a relevance score based on preferred languages
func (r *Repository) calculateRelevance(preferredLanguages []string) {
	score := 0

	// Base score from stars (logarithmic scale)
	if r.Repository.StargazersCount != nil {
		stars := *r.Repository.StargazersCount
		if stars > 0 {
			score += min(100, stars/10)
		}
	}

	// Recent activity bonus
	if r.Repository.UpdatedAt != nil && r.Repository.UpdatedAt.After(time.Now().AddDate(0, -1, 0)) {
		score += 20
	}

	// Language preference bonus
	if r.Repository.Language != nil {
		repoLang := strings.ToLower(*r.Repository.Language)
		for _, prefLang := range preferredLanguages {
			if strings.ToLower(prefLang) == repoLang {
				score += 50
				break
			}
		}
	}

	r.RelevanceScore = score
}

// calculateDifficulty estimates issue difficulty based on labels and content
func (i *Issue) calculateDifficulty() {
	score := 50 // default intermediate

	for _, label := range i.Issue.Labels {
		labelName := strings.ToLower(*label.Name)
		switch {
		case strings.Contains(labelName, "good first issue") ||
			strings.Contains(labelName, "beginner") ||
			strings.Contains(labelName, "easy"):
			score = 20
		case strings.Contains(labelName, "help wanted"):
			score = min(score, 40)
		case strings.Contains(labelName, "hard") ||
			strings.Contains(labelName, "difficult") ||
			strings.Contains(labelName, "expert"):
			score = 80
		case strings.Contains(labelName, "bug"):
			score += 10
		case strings.Contains(labelName, "feature"):
			score += 5
		}
	}

	// Adjust based on comments count (more comments might mean more complex)
	if i.Issue.Comments != nil {
		comments := *i.Issue.Comments
		if comments > 10 {
			score += 15
		} else if comments < 3 {
			score -= 10
		}
	}

	i.DifficultyScore = max(10, min(100, score))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
