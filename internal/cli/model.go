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
)

// keyMap defines keybindings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
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
	repos []*github.Repository
}

type issuesLoadedMsg struct {
	issues []*github.Issue
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
	selectedRepo  *github.Repository
	selectedIssue *github.Issue
	
	// UI state
	loading       bool
	error         error
	width         int
	height        int
	
	// Components
	repoList      list.Model
	issueList     list.Model
	
	keys          keyMap
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
	stars := ""
	if i.repo.Repository.StargazersCount != nil {
		stars = fmt.Sprintf("⭐ %d", *i.repo.Repository.StargazersCount)
	}
	
	lang := ""
	if i.repo.Repository.Language != nil {
		lang = fmt.Sprintf("• %s", *i.repo.Repository.Language)
	}
	
	score := fmt.Sprintf("[Score: %d]", i.repo.RelevanceScore)
	
	desc := ""
	if i.repo.Repository.Description != nil {
		desc = *i.repo.Repository.Description
		if len(desc) > 60 {
			desc = desc[:60] + "..."
		}
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
	// Create repository list
	repoList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	repoList.Title = "Hacktoberfest Repositories"
	repoList.SetShowStatusBar(true)
	repoList.SetFilteringEnabled(true)
	repoList.SetShowHelp(true)
	
	// Create issue list
	issueList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	issueList.Title = "Repository Issues"
	issueList.SetShowStatusBar(true)
	issueList.SetFilteringEnabled(true)
	issueList.SetShowHelp(true)
	
	return Model{
		config:        cfg,
		github:        github.NewClient(cfg.GitHubToken),
		currentScreen: welcomeScreen,
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
			
		case key.Matches(msg, m.keys.Enter):
			return m.handleEnter()
			
		case key.Matches(msg, m.keys.Refresh):
			return m.handleRefresh()
		}
		
	case reposLoadedMsg:
		m.loading = false
		m.repos = msg.repos
		
		// Convert to list items
		items := make([]list.Item, len(msg.repos))
		for i, repo := range msg.repos {
			items[i] = repoItem{repo: repo}
		}
		
		m.repoList.SetItems(items)
		m.currentScreen = repoListScreen
		
	case issuesLoadedMsg:
		m.loading = false
		m.issues = msg.issues
		
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
		return m, m.loadRepositories()
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
	return func() tea.Msg {
		repos, err := m.github.SearchHacktoberfestRepos(20, m.config.PreferredLanguages, m.config.MaxRepos)
		if err != nil {
			return errorMsg{err: err}
		}
		return reposLoadedMsg{repos: repos}
	}
}

func (m Model) loadIssues(repo *github.Repository) tea.Cmd {
	return func() tea.Msg {
		issues, err := m.github.GetRepositoryIssues(
			*repo.Repository.Owner.Login,
			*repo.Repository.Name,
			[]string{"hacktoberfest"},
			m.config.MaxIssuesPerRepo,
		)
		if err != nil {
			return errorMsg{err: err}
		}
		return issuesLoadedMsg{issues: issues}
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
	
	return m.repoList.View()
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
			FooterStyle.Render("Q: Back • R: Retry"),
		)
	}
	
	if len(m.issues) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderHeader("No Issues Found"),
			"",
			RenderError("No suitable issues found in this repository."),
			RenderStatus("This repository might not have beginner-friendly issues."),
			"",
			FooterStyle.Render("Q: Back • R: Refresh"),
		)
	}
	
	return m.issueList.View()
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