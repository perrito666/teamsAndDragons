package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSecondary = lipgloss.Color("#6366F1") // Indigo
	colorPositive  = lipgloss.Color("#22C55E") // Green
	colorNegative  = lipgloss.Color("#EF4444") // Red
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorBorder    = lipgloss.Color("#374151")
	colorHighlight = lipgloss.Color("#E0E7FF")
)

// Styles defines all application styles.
type Styles struct {
	// Layout
	App           lipgloss.Style
	Header        lipgloss.Style
	Content       lipgloss.Style
	Footer        lipgloss.Style
	Breadcrumb    lipgloss.Style
	BreadcrumbSep lipgloss.Style

	// List items
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemTitle    lipgloss.Style
	ListItemDesc     lipgloss.Style

	// Cards
	Card       lipgloss.Style
	CardTitle  lipgloss.Style
	CardBody   lipgloss.Style
	CardFooter lipgloss.Style

	// Status badges
	StatusInProgress lipgloss.Style
	StatusCompleted  lipgloss.Style
	StatusAchieved   lipgloss.Style
	StatusPending    lipgloss.Style

	// Impact badges
	ImpactPositive lipgloss.Style
	ImpactNegative lipgloss.Style

	// Text styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Muted       lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Help        lipgloss.Style
	Prompt      lipgloss.Style
	Placeholder lipgloss.Style

	// Progress bar
	ProgressBar      lipgloss.Style
	ProgressBarFill  lipgloss.Style
	ProgressBarEmpty lipgloss.Style

	// Input
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	TextArea    lipgloss.Style
	TextAreaFocus lipgloss.Style

	// Table
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableCell   lipgloss.Style

	// Tally
	TallyPositive lipgloss.Style
	TallyNegative lipgloss.Style
	TallyNet      lipgloss.Style
}

// DefaultStyles returns the default style configuration.
func DefaultStyles() Styles {
	return Styles{
		// Layout
		App: lipgloss.NewStyle().
			Padding(1, 2),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1),

		Content: lipgloss.NewStyle().
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(colorSecondary),

		BreadcrumbSep: lipgloss.NewStyle().
			Foreground(colorMuted).
			SetString(" > "),

		// List items
		ListItem: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(0),

		ListItemSelected: lipgloss.NewStyle().
			Padding(0, 2).
			Background(colorHighlight).
			Foreground(lipgloss.Color("#1F2937")).
			Bold(true),

		ListItemTitle: lipgloss.NewStyle().
			Bold(true),

		ListItemDesc: lipgloss.NewStyle().
			Foreground(colorMuted),

		// Cards
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginBottom(1),

		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1),

		CardBody: lipgloss.NewStyle(),

		CardFooter: lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1),

		// Status badges
		StatusInProgress: lipgloss.NewStyle().
			Background(lipgloss.Color("#FEF3C7")).
			Foreground(lipgloss.Color("#92400E")).
			Padding(0, 1).
			Bold(true),

		StatusCompleted: lipgloss.NewStyle().
			Background(lipgloss.Color("#D1FAE5")).
			Foreground(lipgloss.Color("#065F46")).
			Padding(0, 1).
			Bold(true),

		StatusAchieved: lipgloss.NewStyle().
			Background(lipgloss.Color("#DBEAFE")).
			Foreground(lipgloss.Color("#1E40AF")).
			Padding(0, 1).
			Bold(true),

		StatusPending: lipgloss.NewStyle().
			Background(lipgloss.Color("#F3F4F6")).
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1),

		// Impact badges
		ImpactPositive: lipgloss.NewStyle().
			Foreground(colorPositive).
			Bold(true),

		ImpactNegative: lipgloss.NewStyle().
			Foreground(colorNegative).
			Bold(true),

		// Text styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary),

		Subtitle: lipgloss.NewStyle().
			Foreground(colorSecondary),

		Label: lipgloss.NewStyle().
			Foreground(colorMuted),

		Value: lipgloss.NewStyle().
			Bold(true),

		Muted: lipgloss.NewStyle().
			Foreground(colorMuted),

		Error: lipgloss.NewStyle().
			Foreground(colorNegative).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(colorPositive).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(colorWarning),

		Help: lipgloss.NewStyle().
			Foreground(colorMuted),

		Prompt: lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true),

		Placeholder: lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true),

		// Progress bar
		ProgressBar: lipgloss.NewStyle().
			Width(20),

		ProgressBarFill: lipgloss.NewStyle().
			Background(colorPositive).
			Foreground(colorPositive),

		ProgressBarEmpty: lipgloss.NewStyle().
			Background(lipgloss.Color("#E5E7EB")).
			Foreground(lipgloss.Color("#E5E7EB")),

		// Input
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1),

		InputFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1),

		TextArea: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1),

		TextAreaFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1),

		// Table
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderBottom(true).
			BorderForeground(colorBorder),

		TableRow: lipgloss.NewStyle().
			Padding(0, 1),

		TableCell: lipgloss.NewStyle().
			Padding(0, 2),

		// Tally
		TallyPositive: lipgloss.NewStyle().
			Foreground(colorPositive).
			Bold(true),

		TallyNegative: lipgloss.NewStyle().
			Foreground(colorNegative).
			Bold(true),

		TallyNet: lipgloss.NewStyle().
			Bold(true),
	}
}

// RenderProgressBar renders a progress bar.
func (s Styles) RenderProgressBar(percent int, width int) string {
	if width <= 0 {
		width = 20
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * width) / 100
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += s.ProgressBarFill.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += s.ProgressBarEmpty.Render("░")
	}

	return bar
}
