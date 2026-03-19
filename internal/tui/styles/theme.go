package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/config"
)

var (
	Primary   = lipgloss.Color("#7aa2f7")
	Secondary = lipgloss.Color("#bb9af7")
	Success   = lipgloss.Color("#9ece6a")
	Warning   = lipgloss.Color("#e0af68")
	Error     = lipgloss.Color("#f7768e")
	Muted     = lipgloss.Color("#565f89")
	Text      = lipgloss.Color("#c0caf5")
	BgDark    = lipgloss.Color("#1a1b26")
	BgOverlay = lipgloss.Color("#24283b")

	Bold     = lipgloss.NewStyle().Bold(true)
	Faint    = lipgloss.NewStyle().Foreground(Muted)
	ErrorTxt = lipgloss.NewStyle().Foreground(Error)
	SuccTxt  = lipgloss.NewStyle().Foreground(Success)
	WarnTxt  = lipgloss.NewStyle().Foreground(Warning)

	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary)

	InactiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted)

	TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 2)

	TabInactive = lipgloss.NewStyle().
			Foreground(Muted).
			Padding(0, 2)

	StatusBar = lipgloss.NewStyle().
			Foreground(Text).
			Background(lipgloss.Color("#24283b")).
			Padding(0, 1)
)

func ApplyTheme(theme config.ThemeConfig) {
	c := theme.Colors
	if c.Primary != "" {
		Primary = lipgloss.Color(c.Primary)
	}
	if c.Secondary != "" {
		Secondary = lipgloss.Color(c.Secondary)
	}
	if c.Success != "" {
		Success = lipgloss.Color(c.Success)
	}
	if c.Warning != "" {
		Warning = lipgloss.Color(c.Warning)
	}
	if c.Error != "" {
		Error = lipgloss.Color(c.Error)
	}

	Faint = lipgloss.NewStyle().Foreground(Muted)
	ErrorTxt = lipgloss.NewStyle().Foreground(Error)
	SuccTxt = lipgloss.NewStyle().Foreground(Success)
	WarnTxt = lipgloss.NewStyle().Foreground(Warning)
	ActiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Primary)
	InactiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Muted)
	TabActive = lipgloss.NewStyle().Bold(true).Foreground(Primary).Padding(0, 2)
	TabInactive = lipgloss.NewStyle().Foreground(Muted).Padding(0, 2)
}
