package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hacktober/internal/config"
	"hacktober/internal/github"
	"hacktober/internal/logger"
)

// keyMap defines keybindings
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Refresh key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Refresh, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "previous page"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next page"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("q/esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

// Screen states
type screen int

const (
	welcomeScreen screen = iota
	repoListScreen
	issueListScreen
	issueDetailScreen
)

// Messages for communication between components
type reposLoadedMsg struct {
	repos        []*github.Repository
	totalRepoCnt int
	currentPage  int
	hasMore      bool
	resetToFirst bool // true for right/next page, false for left/prev page
}

type issuesLoadedMsg struct {
	issues     []*github.Issue
	labelStats map[string]int
}

type repoSelectedMsg struct {
	repo *github.Repository
}

type issueSelectedMsg struct {
	issue *github.Issue
}

type errorMsg struct {
	err error
}

// Main model
type Model struct {
	config        *config.Config
	github        *github.Client
	currentScreen screen

	// Data
	repos         []*github.Repository
	issues        []*github.Issue
	labelStats    map[string]int
	selectedRepo  *github.Repository
	selectedIssue *github.Issue

	// Pagination state
	currentPage  int
	totalRepos   int
	hasMorePages bool

	// UI state
	loading bool
	error   error
	width   int
	height  int

	// Components
	repoList  list.Model
	issueList list.Model

	keys keyMap
}

// Repository list item for bubbles list
type repoItem struct {
	repo *github.Repository
}

func (i repoItem) FilterValue() string {
	return fmt.Sprintf("%s/%s %s",
		*i.repo.Repository.Owner.Login,
		*i.repo.Repository.Name,
		*i.repo.Repository.Description)
}

func (i repoItem) Title() string {
	return fmt.Sprintf("%s/%s", *i.repo.Repository.Owner.Login, *i.repo.Repository.Name)
}

func (i repoItem) Description() string {
	// First line: stars, language, and score
	stars := ""
	if i.repo.Repository.StargazersCount != nil {
		stars = fmt.Sprintf("⭐ %d", *i.repo.Repository.StargazersCount)
	}

	lang := ""
	if i.repo.Repository.Language != nil {
		lang = fmt.Sprintf("• %s", *i.repo.Repository.Language)
	}

	score := fmt.Sprintf("[Score: %d]", i.repo.RelevanceScore)

	// Second line: repository description
	desc := ""
	if i.repo.Repository.Description != nil && *i.repo.Repository.Description != "" {
		desc = *i.repo.Repository.Description
		// Truncate description if too long, but be more generous with space
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		desc = "📝 " + desc
	} else {
		desc = "📝 No description available"
	}

	return fmt.Sprintf("%s %s %s\n%s", stars, lang, score, desc)
}

// Issue list item for bubbles list
type issueItem struct {
	issue *github.Issue
}

func (i issueItem) FilterValue() string {
	return fmt.Sprintf("#%d %s", *i.issue.Issue.Number, *i.issue.Issue.Title)
}

func (i issueItem) Title() string {
	difficulty := ""
	switch {
	case i.issue.DifficultyScore <= 30:
		difficulty = "[Easy]"
	case i.issue.DifficultyScore <= 60:
		difficulty = "[Medium]"
	case i.issue.DifficultyScore <= 80:
		difficulty = "[Hard]"
	default:
		difficulty = "[Expert]"
	}

	return fmt.Sprintf("#%d: %s %s", *i.issue.Issue.Number, *i.issue.Issue.Title, difficulty)
}

func (i issueItem) Description() string {
	comments := ""
	if i.issue.Issue.Comments != nil && *i.issue.Issue.Comments > 0 {
		comments = fmt.Sprintf("💬 %d", *i.issue.Issue.Comments)
	}

	created := i.issue.Issue.CreatedAt.Format("Jan 2, 2006")

	labels := []string{}
	for _, label := range i.issue.Issue.Labels {
		labels = append(labels, *label.Name)
		if len(labels) >= 3 { // Limit to prevent overflow
			break
		}
	}
	labelStr := strings.Join(labels, ", ")

	return fmt.Sprintf("%s • Created: %s\nLabels: %s", comments, created, labelStr)
}

// Initialize the model
func NewModel(cfg *config.Config) Model {
	// Create repository list with custom delegate for better description display
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(4)  // Allow more space for descriptions (title + description lines)
	delegate.SetSpacing(0) // No extra spacing, let content determine spacing

	repoList := list.New([]list.Item{}, delegate, 0, 0)
	repoList.Title = "Hacktoberfest Repositories"
	repoList.SetShowStatusBar(false) // Hide the built-in item count
	repoList.SetFilteringEnabled(true)
	repoList.SetShowHelp(true)

	// Create issue list
	issueDelegate := list.NewDefaultDelegate()
	issueDelegate.SetHeight(3) // Allow for issue title + details
	issueDelegate.SetSpacing(0)

	issueList := list.New([]list.Item{}, issueDelegate, 0, 0)
	issueList.Title = "Repository Issues"
	issueList.SetShowStatusBar(true)
	issueList.SetFilteringEnabled(true)
	issueList.SetShowHelp(true)

	return Model{
		config:        cfg,
		github:        github.NewClient(cfg.GitHubToken),
		currentScreen: welcomeScreen,
		currentPage:   1,
		repoList:      repoList,
		issueList:     issueList,
		keys:          keys,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.repoList.SetWidth(msg.Width)
		m.repoList.SetHeight(msg.Height - 10) // Leave space for header/footer
		m.issueList.SetWidth(msg.Width)
		m.issueList.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		if m.loading {
			// Don't process keys while loading
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Back):
			return m.handleBack()

		case key.Matches(msg, m.keys.Left):
			if m.currentScreen == repoListScreen && m.currentPage > 1 {
				m.loading = true
				m.currentPage--
				return m, m.loadRepositoriesPageWithDirection(m.currentPage, false) // false = go to last item
			}

		case key.Matches(msg, m.keys.Right):
			if m.currentScreen == repoListScreen && m.hasMorePages {
				m.loading = true
				m.currentPage++
				return m, m.loadRepositoriesPageWithDirection(m.currentPage, true) // true = go to first item
			}

		case key.Matches(msg, m.keys.Enter):
			return m.handleEnter()

		case key.Matches(msg, m.keys.Refresh):
			return m.handleRefresh()
		}

	case reposLoadedMsg:
		m.loading = false
		m.repos = msg.repos
		m.currentPage = msg.currentPage
		m.totalRepos = msg.totalRepoCnt
		m.hasMorePages = msg.hasMore

		// Update title with just total count, no page details
		m.repoList.Title = fmt.Sprintf("Hacktoberfest Repositories (~%d total found)", msg.totalRepoCnt)

		// Convert to list items
		items := make([]list.Item, len(msg.repos))
		for i, repo := range msg.repos {
			items[i] = repoItem{repo: repo}
		}

		m.repoList.SetItems(items)

		// Set cursor position based on navigation direction
		if len(items) > 0 {
			if msg.resetToFirst {
				m.repoList.Select(0) // Go to first item
			} else {
				m.repoList.Select(len(items) - 1) // Go to last item
			}
		}

		m.currentScreen = repoListScreen

	case issuesLoadedMsg:
		m.loading = false
		m.issues = msg.issues
		m.labelStats = msg.labelStats

		// Convert to list items
		items := make([]list.Item, len(msg.issues))
		for i, issue := range msg.issues {
			items[i] = issueItem{issue: issue}
		}

		m.issueList.SetItems(items)
		m.currentScreen = issueListScreen

	case errorMsg:
		m.loading = false
		m.error = msg.err
	}

	// Update the current list component
	switch m.currentScreen {
	case repoListScreen:
		m.repoList, cmd = m.repoList.Update(msg)
		cmds = append(cmds, cmd)
	case issueListScreen:
		m.issueList, cmd = m.issueList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleBack() (Model, tea.Cmd) {
	switch m.currentScreen {
	case repoListScreen:
		m.currentScreen = welcomeScreen
	case issueListScreen:
		m.currentScreen = repoListScreen
	case issueDetailScreen:
		m.currentScreen = issueListScreen
	}
	return m, nil
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.currentScreen {
	case welcomeScreen:
		// Start loading repositories
		m.loading = true
		return m, m.loadRepositories()

	case repoListScreen:
		if len(m.repoList.Items()) == 0 {
			return m, nil
		}

		// Get selected repository
		if selectedItem, ok := m.repoList.SelectedItem().(repoItem); ok {
			m.selectedRepo = selectedItem.repo
			m.loading = true
			return m, m.loadIssues(selectedItem.repo)
		}

	case issueListScreen:
		if len(m.issueList.Items()) == 0 {
			return m, nil
		}

		// Get selected issue
		if selectedItem, ok := m.issueList.SelectedItem().(issueItem); ok {
			m.selectedIssue = selectedItem.issue
			m.currentScreen = issueDetailScreen
		}
	}

	return m, nil
}

func (m Model) handleRefresh() (Model, tea.Cmd) {
	switch m.currentScreen {
	case repoListScreen:
		m.loading = true
		return m, m.loadRepositoriesPage(m.currentPage)
	case issueListScreen:
		if m.selectedRepo != nil {
			m.loading = true
			return m, m.loadIssues(m.selectedRepo)
		}
	}
	return m, nil
}

// Commands for async operations
func (m Model) loadRepositories() tea.Cmd {
	return m.loadRepositoriesPageWithDirection(1, true) // Start at first item on initial load
}

func (m Model) loadRepositoriesPage(page int) tea.Cmd {
	return m.loadRepositoriesPageWithDirection(page, true) // Default to first item
}

func (m Model) loadRepositoriesPageWithDirection(page int, resetToFirst bool) tea.Cmd {
	return func() tea.Msg {
		logger.Info(fmt.Sprintf("Loading repositories page %d via CLI command - languages: %v, max: %d",
			page, m.config.PreferredLanguages, m.config.MaxRepos))

		repos, total, err := m.github.SearchHacktoberfestReposWithPage(20, m.config.PreferredLanguages, m.config.MaxRepos, page)
		if err != nil {
			logger.ErrorWithErr("Repository loading failed in CLI", err)
			return errorMsg{err: err}
		}

		// Check if there are more pages
		hasMore := len(repos) == m.config.MaxRepos && (page*m.config.MaxRepos) < total

		logger.Info(fmt.Sprintf("Repositories page %d loaded successfully in CLI: %d repos returned (global total ~%d), hasMore: %t",
			page, len(repos), total, hasMore))

		return reposLoadedMsg{
			repos:        repos,
			totalRepoCnt: total,
			currentPage:  page,
			hasMore:      hasMore,
			resetToFirst: resetToFirst,
		}
	}
}

func (m Model) loadIssues(repo *github.Repository) tea.Cmd {
	return func() tea.Msg {
		repoName := fmt.Sprintf("%s/%s", *repo.Repository.Owner.Login, *repo.Repository.Name)

		logger.Info(fmt.Sprintf("Loading issues for %s via CLI command, max: %d",
			repoName, m.config.MaxIssuesPerRepo))

		issueStats, err := m.github.GetRepositoryIssues(
			*repo.Repository.Owner.Login,
			*repo.Repository.Name,
			[]string{"hacktoberfest"},
			m.config.MaxIssuesPerRepo,
		)
		if err != nil {
			logger.ErrorWithErr("Issue loading failed in CLI", err)
			return errorMsg{err: err}
		}

		logger.Info(fmt.Sprintf("Issues loaded successfully for %s: %d issues found with %d unique labels",
			repoName, issueStats.TotalIssues, len(issueStats.LabelCounts)))

		return issuesLoadedMsg{issues: issueStats.Issues, labelStats: issueStats.LabelCounts}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentScreen {
	case welcomeScreen:
		return m.welcomeView()
	case repoListScreen:
		return m.repoListView()
	case issueListScreen:
		return m.issueListView()
	case issueDetailScreen:
		return m.issueDetailView()
	}

	return "Unknown screen"
}

func (m Model) welcomeView() string {
	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("Searching Repositories"),
			"",
			RenderStatus("Searching for Hacktoberfest repositories..."),
			RenderStatus("This may take a few moments..."),
			"",
			MetaStyle.Render("Press Ctrl+C to cancel"),
		)
	}

	content := []string{
		RenderHeader("Hacktoberfest Repository & Issue Explorer"),
		"",
		RenderSubHeader("Welcome!"),
		ContentStyle.Render("This tool helps you find relevant Hacktoberfest repositories and issues"),
		ContentStyle.Render("tailored to your skills and interests."),
		"",
		RenderSubHeader("Your Configuration"),
		ContentStyle.Render(fmt.Sprintf("Languages: %s", strings.Join(m.config.PreferredLanguages, ", "))),
		ContentStyle.Render(fmt.Sprintf("Skill Level: %s", m.config.SkillLevel)),
		ContentStyle.Render(fmt.Sprintf("Max Repositories: %d", m.config.MaxRepos)),
		"",
		SuccessStyle.Render("Press ENTER to start searching for repositories!"),
		"",
		FooterStyle.Render("Enter: Start • Ctrl+C: Quit"),
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m Model) repoListView() string {
	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("Loading Repositories"),
			"",
			RenderStatus("Please wait..."),
		)
	}

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("Error"),
			"",
			RenderError(fmt.Sprintf("Failed to load repositories: %v", m.error)),
			"",
			RenderStatus("Check logs for details:"),
			MetaStyle.Render(logger.GetLogLocation()),
			"",
			FooterStyle.Render("Q: Back • R: Retry"),
		)
	}

	if len(m.repos) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("No Repositories Found"),
			"",
			RenderError("No repositories found matching your criteria."),
			RenderStatus("Try adjusting your language preferences."),
			"",
			FooterStyle.Render("Q: Back • R: Refresh"),
		)
	}

	// m.repoList.Title already updated with counts; add pagination info and controls
	listView := m.repoList.View()

	// Build pagination info
	var controls []string

	// Add current page info
	totalPages := (m.totalRepos + m.config.MaxRepos - 1) / m.config.MaxRepos // ceil division
	controls = append(controls, fmt.Sprintf("Page %d/%d", m.currentPage, totalPages))

	if m.currentPage > 1 {
		controls = append(controls, "← Previous (left)")
	}
	if m.hasMorePages {
		controls = append(controls, "Next → (right)")
	}
	controls = append(controls, "Type to filter", "R: Refresh", "Q: Back")

	controlText := strings.Join(controls, " • ")
	info := MetaStyle.Render(controlText)

	return lipgloss.JoinVertical(lipgloss.Left, listView, info)
}

func (m Model) issueListView() string {
	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("Loading Issues"),
			"",
			RenderStatus("Please wait..."),
		)
	}

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("Error"),
			"",
			RenderError(fmt.Sprintf("Failed to load issues: %v", m.error)),
			"",
			RenderStatus("Check logs for details:"),
			MetaStyle.Render(logger.GetLogLocation()),
			"",
			FooterStyle.Render("Q: Back • R: Retry"),
		)
	}

	if len(m.issues) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("No Issues Found"),
			"",
			RenderError("No open issues found in this repository."),
			RenderStatus("This repository might not have any open issues."),
			"",
			FooterStyle.Render("Q: Back • R: Refresh"),
		)
	}

	// Create header with repository info and label statistics
	repoName := "Unknown"
	if m.selectedRepo != nil {
		repoName = fmt.Sprintf("%s/%s", *m.selectedRepo.Owner.Login, *m.selectedRepo.Name)
	}

	header := RenderHeader(fmt.Sprintf("Issues in %s", repoName))

	// Build label statistics display
	var labelLines []string
	if len(m.labelStats) > 0 {
		labelLines = append(labelLines, RenderStatus(fmt.Sprintf("Found %d issues with %d unique labels:",
			len(m.issues), len(m.labelStats))))

		// Sort labels by count for better display
		type labelCount struct {
			name  string
			count int
		}
		var sortedLabels []labelCount
		for label, count := range m.labelStats {
			sortedLabels = append(sortedLabels, labelCount{label, count})
		}

		// Sort by count (descending)
		for i := 0; i < len(sortedLabels)-1; i++ {
			for j := i + 1; j < len(sortedLabels); j++ {
				if sortedLabels[i].count < sortedLabels[j].count {
					sortedLabels[i], sortedLabels[j] = sortedLabels[j], sortedLabels[i]
				}
			}
		}

		// Show top 10 labels
		maxLabels := len(sortedLabels)
		if maxLabels > 10 {
			maxLabels = 10
		}

		var labelStrs []string
		for i := 0; i < maxLabels; i++ {
			labelStrs = append(labelStrs, fmt.Sprintf("%s (%d)", sortedLabels[i].name, sortedLabels[i].count))
		}

		labelLines = append(labelLines, MetaStyle.Render(strings.Join(labelStrs, " • ")))
		if len(sortedLabels) > 10 {
			labelLines = append(labelLines, MetaStyle.Render(fmt.Sprintf("... and %d more labels", len(sortedLabels)-10)))
		}
	} else {
		labelLines = append(labelLines, RenderStatus(fmt.Sprintf("Found %d issues", len(m.issues))))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		strings.Join(labelLines, "\n"),
		"",
		m.issueList.View(),
	)
}

func (m Model) issueDetailView() string {
	if m.selectedIssue == nil || m.selectedRepo == nil {
		return "No issue selected"
	}

	issue := m.selectedIssue
	repo := m.selectedRepo

	content := []string{
		RenderHeader(fmt.Sprintf("Issue #%d: %s/%s",
			*issue.Issue.Number,
			*repo.Repository.Owner.Login,
			*repo.Repository.Name)),
		"",
		// Title
		lipgloss.NewStyle().Bold(true).Foreground(Text).Render(*issue.Issue.Title),
		"",
		// Metadata
		RenderSubHeader("Details"),
	}

	// Author
	content = append(content, ContentStyle.Render(fmt.Sprintf("Author: %s", *issue.Issue.User.Login)))

	// Created date
	content = append(content, ContentStyle.Render(fmt.Sprintf("Created: %s",
		issue.Issue.CreatedAt.Format("January 2, 2006 at 15:04"))))

	// Comments
	if issue.Issue.Comments != nil {
		content = append(content, ContentStyle.Render(fmt.Sprintf("Comments: %d", *issue.Issue.Comments)))
	}

	// Difficulty
	difficulty := RenderDifficulty(issue.DifficultyScore)
	content = append(content, ContentStyle.Render(fmt.Sprintf("Difficulty: %s (%d/100)",
		difficulty, issue.DifficultyScore)))

	// Labels
	if len(issue.Issue.Labels) > 0 {
		labels := []string{}
		for _, label := range issue.Issue.Labels {
			labels = append(labels, *label.Name)
		}
		content = append(content, ContentStyle.Render(fmt.Sprintf("Labels: %s",
			strings.Join(labels, ", "))))
	}

	// URL
	content = append(content, ContentStyle.Render(fmt.Sprintf("URL: %s", *issue.Issue.HTMLURL)))
	content = append(content, "")

	// Description
	if issue.Issue.Body != nil && *issue.Issue.Body != "" {
		content = append(content, RenderSubHeader("Description"))

		// Wrap the body text
		body := *issue.Issue.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		content = append(content, DescriptionStyle.Render(body))
	}

	content = append(content, "")
	content = append(content, FooterStyle.Render("Q: Back"))

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}
