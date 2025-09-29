package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Screen represents different CLI screens
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenRepoSearch
	ScreenRepoList
	ScreenIssueList
	ScreenIssueDetail
)

// KeyCode represents keyboard input
type KeyCode int

const (
	KeyUp KeyCode = iota
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeyQ
	KeyR
	KeyOther
)

// Input handles keyboard input
type Input struct {
	oldState *term.State
}

// NewInput creates a new input handler
func NewInput() (*Input, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &Input{oldState: oldState}, nil
}

// Close restores terminal state
func (i *Input) Close() {
	if i.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), i.oldState)
	}
}

// ReadKey reads a single key press
func (i *Input) ReadKey() KeyCode {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return KeyOther
	}

	if n == 1 {
		switch buf[0] {
		case 13: // Enter
			return KeyEnter
		case 27: // Escape
			return KeyEsc
		case 'q', 'Q':
			return KeyQ
		case 'r', 'R':
			return KeyR
		default:
			return KeyOther
		}
	}

	if n == 3 && buf[0] == 27 && buf[1] == 91 { // Arrow keys
		switch buf[2] {
		case 65: // Up
			return KeyUp
		case 66: // Down
			return KeyDown
		case 67: // Right
			return KeyRight
		case 68: // Left
			return KeyLeft
		}
	}

	return KeyOther
}

// Display handles screen rendering
type Display struct {
	width  int
	height int
}

// NewDisplay creates a new display handler
func NewDisplay() *Display {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24 // fallback
	}

	return &Display{
		width:  width,
		height: height,
	}
}

// Clear clears the screen
func (d *Display) Clear() {
	fmt.Print("\033[2J\033[H")
}

// MoveCursor moves cursor to position
func (d *Display) MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y+1, x+1)
}

// SetColor sets text color (ANSI codes)
func (d *Display) SetColor(color string) {
	colors := map[string]string{
		"reset":   "\033[0m",
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
		"bold":    "\033[1m",
		"dim":     "\033[2m",
	}

	if code, ok := colors[color]; ok {
		fmt.Print(code)
	}
}

// PrintLine prints a line with word wrapping
func (d *Display) PrintLine(text string, maxWidth int) {
	if len(text) <= maxWidth {
		fmt.Println(text)
		return
	}

	words := strings.Fields(text)
	line := ""

	for _, word := range words {
		if len(line)+len(word)+1 > maxWidth {
			fmt.Println(line)
			line = word
		} else {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		}
	}

	if line != "" {
		fmt.Println(line)
	}
}

// PrintHeader prints a section header
func (d *Display) PrintHeader(title string) {
	d.SetColor("bold")
	d.SetColor("cyan")
	fmt.Printf("═══ %s ", title)
	for i := len(title) + 5; i < d.width; i++ {
		fmt.Print("═")
	}
	fmt.Println()
	d.SetColor("reset")
}

// PrintSubheader prints a subsection header
func (d *Display) PrintSubheader(title string) {
	d.SetColor("bold")
	d.SetColor("yellow")
	fmt.Printf("── %s ", title)
	for i := len(title) + 4; i < d.width-5; i++ {
		fmt.Print("─")
	}
	fmt.Println()
	d.SetColor("reset")
}

// PrintSelectedItem prints a selected list item
func (d *Display) PrintSelectedItem(text string) {
	d.SetColor("bold")
	d.SetColor("green")
	fmt.Printf("► %s", text)
	d.SetColor("reset")
	fmt.Println()
}

// PrintItem prints a regular list item
func (d *Display) PrintItem(text string) {
	fmt.Printf("  %s", text)
	fmt.Println()
}

// PrintStatus prints status information
func (d *Display) PrintStatus(text string) {
	d.SetColor("dim")
	d.SetColor("cyan")
	fmt.Printf("ℹ %s", text)
	d.SetColor("reset")
	fmt.Println()
}

// PrintError prints error information
func (d *Display) PrintError(text string) {
	d.SetColor("bold")
	d.SetColor("red")
	fmt.Printf("✗ %s", text)
	d.SetColor("reset")
	fmt.Println()
}

// PrintSuccess prints success information
func (d *Display) PrintSuccess(text string) {
	d.SetColor("bold")
	d.SetColor("green")
	fmt.Printf("✓ %s", text)
	d.SetColor("reset")
	fmt.Println()
}

// GetWidth returns display width
func (d *Display) GetWidth() int {
	return d.width
}

// GetHeight returns display height
func (d *Display) GetHeight() int {
	return d.height
}
