package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme colors
var (
	Primary    = lipgloss.Color("#FF6B35") // Halloween orange
	Secondary  = lipgloss.Color("#F7931E") // Pumpkin orange
	Accent     = lipgloss.Color("#FFD23F") // Yellow
	Success    = lipgloss.Color("#06D6A0") // Green
	Warning    = lipgloss.Color("#F72585") // Pink/Red
	Info       = lipgloss.Color("#4361EE") // Blue
	Muted      = lipgloss.Color("#6C757D") // Gray
	Text       = lipgloss.Color("#FFFFFF") // White
	Background = lipgloss.Color("#000000") // Black
)

// Styles for different UI elements
var (
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Background).
				Background(Primary).
				Bold(true).
				Padding(0, 1).
				MarginRight(1)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1).
			MarginRight(1)

	// Status and info styles
	StatusStyle = lipgloss.NewStyle().
			Foreground(Info).
			Italic(true).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true).
			Padding(0, 1)

	// Metadata styles
	MetaStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Content styles
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1)

	DescriptionStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Padding(0, 2).
				Width(80)

	// Footer styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(Muted).
			Padding(1, 2).
			MarginTop(1)

	// Difficulty styles
	EasyStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	MediumStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	HardStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	ExpertStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	// Pagination styles
	PaginationStyle = lipgloss.NewStyle().
			Foreground(Info).
			Padding(0, 2)

	// Highlight styles for numbers
	NumberStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	StarStyle = lipgloss.NewStyle().
			Foreground(Accent)

	LanguageStyle = lipgloss.NewStyle().
			Foreground(Info)

	DateStyle = lipgloss.NewStyle().
			Foreground(Muted)
)

// Helper functions for consistent styling
func RenderHeader(title string) string {
	return HeaderStyle.Render("🎃 " + title)
}

func RenderSubHeader(title string) string {
	return SubHeaderStyle.Render("── " + title + " ──")
}

func RenderSelectedItem(text string) string {
	return SelectedItemStyle.Render("► " + text)
}

func RenderNormalItem(text string) string {
	return NormalItemStyle.Render("  " + text)
}

func RenderStatus(text string) string {
	return StatusStyle.Render("ℹ " + text)
}

func RenderError(text string) string {
	return ErrorStyle.Render("✗ " + text)
}

func RenderSuccess(text string) string {
	return SuccessStyle.Render("✓ " + text)
}

func RenderDifficulty(score int) string {
	switch {
	case score <= 30:
		return EasyStyle.Render("[Easy]")
	case score <= 60:
		return MediumStyle.Render("[Medium]")
	case score <= 80:
		return HardStyle.Render("[Hard]")
	default:
		return ExpertStyle.Render("[Expert]")
	}
}

func RenderStars(count int) string {
	return StarStyle.Render("⭐ ") + NumberStyle.Render(fmt.Sprintf("%d", count))
}

func RenderLanguage(lang string) string {
	return LanguageStyle.Render("• " + lang)
}

func RenderRelevanceScore(score int) string {
	return LabelStyle.Render("[Score: ") + NumberStyle.Render(fmt.Sprintf("%d", score)) + LabelStyle.Render("]")
}
