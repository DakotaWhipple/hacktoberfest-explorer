package cli

import (
	"fmt"
	"strings"

	"hacktober/internal/config"
	"hacktober/internal/github"
)

// App represents the main CLI application
type App struct {
	config        *config.Config
	github        *github.Client
	display       *Display
	input         *Input
	screen        Screen
	repos         []*github.Repository
	issues        []*github.Issue
	selectedRepo  int
	selectedIssue int
	currentPage   int
	itemsPerPage  int
}

// NewApp creates a new CLI application
func NewApp(cfg *config.Config) *App {
	return &App{
		config:       cfg,
		github:       github.NewClient(cfg.GitHubToken),
		display:      NewDisplay(),
		screen:       ScreenWelcome,
		itemsPerPage: max(5, cfg.MaxRepos/10), // Show reasonable number per page
	}
}

// Run starts the CLI application
func (a *App) Run() error {
	input, err := NewInput()
	if err != nil {
		return fmt.Errorf("failed to initialize input: %w", err)
	}
	defer input.Close()
	a.input = input

	// Main event loop
	for {
		a.render()

		key := a.input.ReadKey()
		if !a.handleInput(key) {
			break
		}
	}

	a.display.Clear()
	return nil
}

// handleInput processes keyboard input
func (a *App) handleInput(key KeyCode) bool {
	switch key {
	case KeyQ, KeyEsc:
		if a.screen == ScreenWelcome {
			return false // Exit application
		}
		a.goBack()
	case KeyEnter:
		a.handleEnter()
	case KeyUp:
		a.handleUp()
	case KeyDown:
		a.handleDown()
	case KeyLeft:
		a.handleLeft()
	case KeyRight:
		a.handleRight()
	case KeyR:
		a.refresh()
	}

	return true
}

// render displays the current screen
func (a *App) render() {
	a.display.Clear()

	switch a.screen {
	case ScreenWelcome:
		a.renderWelcome()
	case ScreenRepoSearch:
		a.renderRepoSearch()
	case ScreenRepoList:
		a.renderRepoList()
	case ScreenIssueList:
		a.renderIssueList()
	case ScreenIssueDetail:
		a.renderIssueDetail()
	}

	a.renderFooter()
}

// renderWelcome shows the welcome screen
func (a *App) renderWelcome() {
	a.display.PrintHeader("🎃 Hacktoberfest Repository & Issue Explorer")
	fmt.Println()

	a.display.PrintSubheader("Welcome!")
	fmt.Println("This tool helps you find relevant Hacktoberfest repositories and issues")
	fmt.Println("tailored to your skills and interests.")
	fmt.Println()

	a.display.PrintStatus("Your Configuration:")
	fmt.Printf("  • Languages: %s\n", strings.Join(a.config.PreferredLanguages, ", "))
	fmt.Printf("  • Skill Level: %s\n", a.config.SkillLevel)
	fmt.Printf("  • Max Repositories: %d\n", a.config.MaxRepos)
	fmt.Println()

	a.display.SetColor("bold")
	a.display.SetColor("green")
	fmt.Println("Press ENTER to start searching for repositories!")
	a.display.SetColor("reset")
}

// renderRepoSearch shows loading screen while searching repositories
func (a *App) renderRepoSearch() {
	a.display.PrintHeader("🔍 Searching Hacktoberfest Repositories")
	fmt.Println()

	a.display.PrintStatus("Searching for repositories with:")
	fmt.Printf("  • Minimum 20 stars\n")
	fmt.Printf("  • Hacktoberfest topic\n")
	fmt.Printf("  • Languages: %s\n", strings.Join(a.config.PreferredLanguages, ", "))
	fmt.Println()

	a.display.SetColor("yellow")
	fmt.Println("⏳ Please wait...")
	a.display.SetColor("reset")
}

// renderRepoList shows the list of repositories
func (a *App) renderRepoList() {
	a.display.PrintHeader("📚 Hacktoberfest Repositories")

	if len(a.repos) == 0 {
		fmt.Println()
		a.display.PrintError("No repositories found matching your criteria.")
		a.display.PrintStatus("Try adjusting your language preferences or search again.")
		return
	}

	start := a.currentPage * a.itemsPerPage
	end := min(start+a.itemsPerPage, len(a.repos))

	fmt.Printf("Showing %d-%d of %d repositories (Page %d)\n\n",
		start+1, end, len(a.repos), a.currentPage+1)

	for i := start; i < end; i++ {
		repo := a.repos[i]

		if i == a.selectedRepo {
			a.display.PrintSelectedItem("")
		} else {
			fmt.Print("  ")
		}

		// Repository name and stars
		a.display.SetColor("bold")
		fmt.Printf("%s/%s", *repo.Repository.Owner.Login, *repo.Repository.Name)
		a.display.SetColor("reset")

		a.display.SetColor("yellow")
		fmt.Printf(" ⭐ %d", *repo.Repository.StargazersCount)
		a.display.SetColor("reset")

		// Language
		if repo.Repository.Language != nil && *repo.Repository.Language != "" {
			a.display.SetColor("cyan")
			fmt.Printf(" • %s", *repo.Repository.Language)
			a.display.SetColor("reset")
		}

		// Relevance score
		a.display.SetColor("green")
		fmt.Printf(" [Score: %d]", repo.RelevanceScore)
		a.display.SetColor("reset")

		fmt.Println()

		// Description
		if repo.Repository.Description != nil && *repo.Repository.Description != "" {
			a.display.SetColor("dim")
			a.display.PrintLine("    "+*repo.Repository.Description, a.display.GetWidth()-8)
			a.display.SetColor("reset")
		}

		// Last updated
		updated := repo.Repository.UpdatedAt.Format("Jan 2, 2006")
		a.display.SetColor("dim")
		fmt.Printf("    Updated: %s\n", updated)
		a.display.SetColor("reset")

		fmt.Println()
	}
}

// renderIssueList shows issues for selected repository
func (a *App) renderIssueList() {
	if a.selectedRepo >= len(a.repos) {
		return
	}

	repo := a.repos[a.selectedRepo]
	a.display.PrintHeader(fmt.Sprintf("🐛 Issues in %s/%s",
		*repo.Repository.Owner.Login, *repo.Repository.Name))

	if len(a.issues) == 0 {
		fmt.Println()
		a.display.PrintError("No suitable issues found in this repository.")
		a.display.PrintStatus("This repository might not have beginner-friendly issues at the moment.")
		return
	}

	start := a.currentPage * a.itemsPerPage
	end := min(start+a.itemsPerPage, len(a.issues))

	fmt.Printf("Showing %d-%d of %d issues (Page %d)\n\n",
		start+1, end, len(a.issues), a.currentPage+1)

	for i := start; i < end; i++ {
		issue := a.issues[i]

		if i == a.selectedIssue {
			a.display.PrintSelectedItem("")
		} else {
			fmt.Print("  ")
		}

		// Issue title and number
		a.display.SetColor("bold")
		fmt.Printf("#%d: %s", *issue.Issue.Number, *issue.Issue.Title)
		a.display.SetColor("reset")

		// Difficulty indicator
		difficulty := a.getDifficultyText(issue.DifficultyScore)
		color := a.getDifficultyColor(issue.DifficultyScore)
		a.display.SetColor(color)
		fmt.Printf(" [%s]", difficulty)
		a.display.SetColor("reset")

		// Comments count
		if issue.Issue.Comments != nil && *issue.Issue.Comments > 0 {
			a.display.SetColor("cyan")
			fmt.Printf(" 💬 %d", *issue.Issue.Comments)
			a.display.SetColor("reset")
		}

		fmt.Println()

		// Labels
		if len(issue.Issue.Labels) > 0 {
			a.display.SetColor("dim")
			fmt.Print("    Labels: ")
			for j, label := range issue.Issue.Labels {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Print(*label.Name)
			}
			fmt.Println()
			a.display.SetColor("reset")
		}

		// Created date
		created := issue.Issue.CreatedAt.Format("Jan 2, 2006")
		a.display.SetColor("dim")
		fmt.Printf("    Created: %s\n", created)
		a.display.SetColor("reset")

		fmt.Println()
	}
}

// renderIssueDetail shows detailed view of selected issue
func (a *App) renderIssueDetail() {
	if a.selectedIssue >= len(a.issues) {
		return
	}

	issue := a.issues[a.selectedIssue]
	repo := a.repos[a.selectedRepo]

	a.display.PrintHeader(fmt.Sprintf("📋 Issue Details: %s/%s #%d",
		*repo.Repository.Owner.Login, *repo.Repository.Name, *issue.Issue.Number))

	fmt.Println()

	// Title
	a.display.SetColor("bold")
	a.display.PrintLine(*issue.Issue.Title, a.display.GetWidth())
	a.display.SetColor("reset")
	fmt.Println()

	// Metadata
	a.display.PrintSubheader("Metadata")
	fmt.Printf("Author: %s\n", *issue.Issue.User.Login)
	fmt.Printf("Created: %s\n", issue.Issue.CreatedAt.Format("January 2, 2006 at 15:04"))
	if issue.Issue.Comments != nil {
		fmt.Printf("Comments: %d\n", *issue.Issue.Comments)
	}

	difficulty := a.getDifficultyText(issue.DifficultyScore)
	color := a.getDifficultyColor(issue.DifficultyScore)
	fmt.Print("Difficulty: ")
	a.display.SetColor(color)
	fmt.Printf("%s (%d/100)\n", difficulty, issue.DifficultyScore)
	a.display.SetColor("reset")

	if len(issue.Issue.Labels) > 0 {
		fmt.Print("Labels: ")
		for i, label := range issue.Issue.Labels {
			if i > 0 {
				fmt.Print(", ")
			}
			a.display.SetColor("cyan")
			fmt.Print(*label.Name)
			a.display.SetColor("reset")
		}
		fmt.Println()
	}

	fmt.Printf("URL: %s\n", *issue.Issue.HTMLURL)
	fmt.Println()

	// Description
	if issue.Issue.Body != nil && *issue.Issue.Body != "" {
		a.display.PrintSubheader("Description")
		a.display.PrintLine(*issue.Issue.Body, a.display.GetWidth()-4)
	}
}

// renderFooter shows navigation help
func (a *App) renderFooter() {
	fmt.Println()
	a.display.SetColor("dim")
	fmt.Print("─")
	for i := 1; i < a.display.GetWidth(); i++ {
		fmt.Print("─")
	}
	fmt.Println()

	switch a.screen {
	case ScreenWelcome:
		fmt.Print("ENTER: Start • Q: Quit")
	case ScreenRepoList:
		fmt.Print("↑↓: Navigate • ENTER: View Issues • R: Refresh • Q: Back")
		if a.canPaginate() {
			fmt.Print(" • ←→: Page")
		}
	case ScreenIssueList:
		fmt.Print("↑↓: Navigate • ENTER: View Details • R: Refresh • Q: Back")
		if a.canPaginate() {
			fmt.Print(" • ←→: Page")
		}
	case ScreenIssueDetail:
		fmt.Print("Q: Back")
	default:
		fmt.Print("Q: Back")
	}

	a.display.SetColor("reset")
	fmt.Println()
}

// Navigation methods
func (a *App) handleEnter() {
	switch a.screen {
	case ScreenWelcome:
		a.screen = ScreenRepoSearch
		a.searchRepositories()
	case ScreenRepoList:
		a.searchIssues()
	case ScreenIssueList:
		a.screen = ScreenIssueDetail
	}
}

func (a *App) handleUp() {
	switch a.screen {
	case ScreenRepoList:
		if a.selectedRepo > 0 {
			a.selectedRepo--
			a.adjustPageForSelection()
		}
	case ScreenIssueList:
		if a.selectedIssue > 0 {
			a.selectedIssue--
			a.adjustPageForSelection()
		}
	}
}

func (a *App) handleDown() {
	switch a.screen {
	case ScreenRepoList:
		if a.selectedRepo < len(a.repos)-1 {
			a.selectedRepo++
			a.adjustPageForSelection()
		}
	case ScreenIssueList:
		if a.selectedIssue < len(a.issues)-1 {
			a.selectedIssue++
			a.adjustPageForSelection()
		}
	}
}

func (a *App) handleLeft() {
	if a.canPaginate() && a.currentPage > 0 {
		a.currentPage--
	}
}

func (a *App) handleRight() {
	if a.canPaginate() {
		maxPage := a.getMaxPage()
		if a.currentPage < maxPage {
			a.currentPage++
		}
	}
}

func (a *App) goBack() {
	switch a.screen {
	case ScreenRepoList:
		a.screen = ScreenWelcome
	case ScreenIssueList:
		a.screen = ScreenRepoList
	case ScreenIssueDetail:
		a.screen = ScreenIssueList
	default:
		a.screen = ScreenWelcome
	}
	a.currentPage = 0
}

func (a *App) refresh() {
	switch a.screen {
	case ScreenRepoList:
		a.searchRepositories()
	case ScreenIssueList:
		a.searchIssues()
	}
}

// Data fetching methods
func (a *App) searchRepositories() {
	repos, err := a.github.SearchHacktoberfestRepos(20, a.config.PreferredLanguages, a.config.MaxRepos)
	if err != nil {
		// Handle error - for now just show empty results
		a.repos = nil
	} else {
		a.repos = repos
	}

	a.selectedRepo = 0
	a.currentPage = 0
	a.screen = ScreenRepoList
}

func (a *App) searchIssues() {
	if a.selectedRepo >= len(a.repos) {
		return
	}

	repo := a.repos[a.selectedRepo]
	issues, err := a.github.GetRepositoryIssues(
		*repo.Repository.Owner.Login,
		*repo.Repository.Name,
		[]string{"hacktoberfest"},
		a.config.MaxIssuesPerRepo,
	)

	if err != nil {
		a.issues = nil
	} else {
		a.issues = issues
	}

	a.selectedIssue = 0
	a.currentPage = 0
	a.screen = ScreenIssueList
}

// Helper methods
func (a *App) canPaginate() bool {
	return a.getMaxPage() > 0
}

func (a *App) getMaxPage() int {
	switch a.screen {
	case ScreenRepoList:
		return (len(a.repos) - 1) / a.itemsPerPage
	case ScreenIssueList:
		return (len(a.issues) - 1) / a.itemsPerPage
	}
	return 0
}

func (a *App) adjustPageForSelection() {
	var selectedIndex int
	switch a.screen {
	case ScreenRepoList:
		selectedIndex = a.selectedRepo
	case ScreenIssueList:
		selectedIndex = a.selectedIssue
	default:
		return
	}

	requiredPage := selectedIndex / a.itemsPerPage
	if requiredPage != a.currentPage {
		a.currentPage = requiredPage
	}
}

func (a *App) getDifficultyText(score int) string {
	switch {
	case score <= 30:
		return "Easy"
	case score <= 60:
		return "Medium"
	case score <= 80:
		return "Hard"
	default:
		return "Expert"
	}
}

func (a *App) getDifficultyColor(score int) string {
	switch {
	case score <= 30:
		return "green"
	case score <= 60:
		return "yellow"
	case score <= 80:
		return "magenta"
	default:
		return "red"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
