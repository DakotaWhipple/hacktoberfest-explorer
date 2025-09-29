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
func (c *Client) SearchHacktoberfestRepos(minStars int, languages []string, maxResults int) ([]*Repository, error) {
	start := time.Now()

	query := fmt.Sprintf("topic:hacktoberfest stars:>=%d sort:stars-desc", minStars)

	// Add language filter if specified
	if len(languages) > 0 {
		langQuery := make([]string, len(languages))
		for i, lang := range languages {
			langQuery[i] = fmt.Sprintf("language:%s", strings.ToLower(lang))
		}
		query += " (" + strings.Join(langQuery, " OR ") + ")"
	}

	logger.Info(fmt.Sprintf("Starting repository search: %s", query))

	opts := &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: min(maxResults, 100), // GitHub API limit
		},
	}

	result, response, err := c.client.Search.Repositories(c.ctx, query, opts)
	duration := time.Since(start)

	if response != nil {
		logger.LogAPIRequest("repositories/search", query, response.StatusCode, duration)
		logger.Debug(fmt.Sprintf("Rate limit remaining: %d, resets at: %v",
			response.Rate.Remaining, response.Rate.Reset.Time))
	}

	if err != nil {
		logger.ErrorWithErr("Failed to search repositories", err)
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	totalFound := 0
	if result.Total != nil {
		totalFound = *result.Total
	}

	logger.Info(fmt.Sprintf("Repository search API completed: %d total found, %d returned, took %v",
		totalFound, len(result.Repositories), duration))

	repos := make([]*Repository, 0, len(result.Repositories))
	for _, repo := range result.Repositories {
		if repo.StargazersCount != nil && *repo.StargazersCount >= minStars {
			r := &Repository{
				Repository: repo,
			}
			r.calculateRelevance(languages)
			repos = append(repos, r)

			// Log each repository found
			logger.Debug(fmt.Sprintf("Repository processed: %s/%s, stars: %d, relevance: %d",
				*repo.Owner.Login, *repo.Name, *repo.StargazersCount, r.RelevanceScore))
		}
	}

	logger.LogRepoSearch(query, totalFound, len(repos), languages)
	logger.Info(fmt.Sprintf("Repository search completed: %d filtered results", len(repos)))

	return repos, nil
}

// GetRepositoryIssues fetches issues for a specific repository
func (c *Client) GetRepositoryIssues(owner, repo string, labels []string, maxResults int) ([]*Issue, error) {
	start := time.Now()
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	logger.Info(fmt.Sprintf("Starting issue search for %s", repoName))

	opts := &github.IssueListByRepoOptions{
		State:  "open",
		Labels: append(labels, "good first issue", "help wanted"),
		Sort:   "updated",
		ListOptions: github.ListOptions{
			PerPage: min(maxResults, 100),
		},
	}

	issues, response, err := c.client.Issues.ListByRepo(c.ctx, owner, repo, opts)
	duration := time.Since(start)

	if response != nil {
		logger.LogAPIRequest("issues/list", repoName, response.StatusCode, duration)
		logger.Debug(fmt.Sprintf("Rate limit remaining: %d, resets at: %v",
			response.Rate.Remaining, response.Rate.Reset.Time))
	}

	if err != nil {
		logger.ErrorWithErr("Failed to fetch repository issues", err)
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	logger.Info(fmt.Sprintf("Issue search API completed for %s: %d issues returned, took %v",
		repoName, len(issues), duration))

	result := make([]*Issue, 0, len(issues))
	for _, issue := range issues {
		// Skip pull requests
		if issue.PullRequestLinks != nil {
			logger.Debug(fmt.Sprintf("Skipping pull request #%d: %s", *issue.Number, *issue.Title))
			continue
		}

		i := &Issue{
			Issue: issue,
		}
		i.calculateDifficulty()
		result = append(result, i)

		// Log each issue found
		logger.Debug(fmt.Sprintf("Issue processed #%d: %s, difficulty: %d",
			*issue.Number, *issue.Title, i.DifficultyScore))
	}

	logger.LogIssueSearch(repoName, len(issues), len(result))
	logger.Info(fmt.Sprintf("Issue search completed for %s: %d filtered results", repoName, len(result)))

	return result, nil
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
