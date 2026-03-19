package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashSuccess
	FlashWarning
	FlashError
)

type Flash struct {
	message string
	level   FlashLevel
	visible bool
}

type FlashClearMsg struct{}

func NewFlash() *Flash {
	return &Flash{}
}

func (f *Flash) Show(msg string, level FlashLevel) tea.Cmd {
	f.message = msg
	f.level = level
	f.visible = true
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return FlashClearMsg{}
	})
}

func (f *Flash) Clear() {
	f.visible = false
	f.message = ""
}

func (f *Flash) View() string {
	if !f.visible {
		return ""
	}

	var style lipgloss.Style
	switch f.level {
	case FlashSuccess:
		style = lipgloss.NewStyle().Foreground(styles.Success)
	case FlashWarning:
		style = lipgloss.NewStyle().Foreground(styles.Warning)
	case FlashError:
		style = lipgloss.NewStyle().Foreground(styles.Error)
	default:
		style = lipgloss.NewStyle().Foreground(styles.Primary)
	}

	return style.Render(f.message)
}
