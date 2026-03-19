package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type Sidebar struct {
	viewport viewport.Model
	Title    string
	Active   bool
	Width    int
	Height   int
	Open     bool
}

func NewSidebar() *Sidebar {
	return &Sidebar{
		viewport: viewport.New(0, 0),
		Open:     true,
	}
}

func (s *Sidebar) SetSize(w, h int) {
	s.Width = w
	s.Height = h
	s.viewport.Width = max(0, w-4)
	s.viewport.Height = max(0, h-4)
}

func (s *Sidebar) SetContent(content string) {
	s.viewport.SetContent(content)
}

func (s *Sidebar) Toggle() {
	s.Open = !s.Open
}

func (s *Sidebar) Update(msg tea.Msg) (*Sidebar, tea.Cmd) {
	if !s.Open {
		return s, nil
	}
	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *Sidebar) View() string {
	if !s.Open {
		return ""
	}

	border := styles.InactiveBorder
	if s.Active {
		border = styles.ActiveBorder
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render(s.Title)
	content := title + "\n" + strings.Repeat("─", max(0, s.Width-4)) + "\n" + s.viewport.View()

	return border.Width(s.Width - 2).Height(s.Height - 2).Render(content)
}
