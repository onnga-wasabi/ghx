package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type LogViewer struct {
	viewport viewport.Model
	content  string
	parsed   *ParsedLogs
	Title    string
	Active   bool
	Width    int
	Height   int
	wordWrap bool
	search   string
	matchIdx int
	matches  []int
}

func NewLogViewer() *LogViewer {
	return &LogViewer{
		viewport: viewport.New(0, 0),
	}
}

func (l *LogViewer) SetContent(raw string) {
	l.content = raw
	l.parsed = ParseLogs(raw)
	l.refreshView()
}

func (l *LogViewer) SetSize(w, h int) {
	l.Width = w
	l.Height = h
	innerW := max(0, w-4)
	innerH := max(0, h-4)
	l.viewport.Width = innerW
	l.viewport.Height = innerH
	l.refreshView()
}

func (l *LogViewer) SetSearch(s string) {
	l.search = s
	l.matchIdx = 0
	l.refreshView()
}

func (l *LogViewer) NextMatch() {
	if len(l.matches) > 0 {
		l.matchIdx = (l.matchIdx + 1) % len(l.matches)
		l.viewport.GotoTop()
		target := max(0, l.matches[l.matchIdx]-l.viewport.Height/2)
		l.viewport.SetYOffset(target)
	}
}

func (l *LogViewer) PrevMatch() {
	if len(l.matches) > 0 {
		l.matchIdx = (l.matchIdx - 1 + len(l.matches)) % len(l.matches)
		l.viewport.GotoTop()
		target := max(0, l.matches[l.matchIdx]-l.viewport.Height/2)
		l.viewport.SetYOffset(target)
	}
}

func (l *LogViewer) ToggleWordWrap() {
	l.wordWrap = !l.wordWrap
	l.refreshView()
}

func (l *LogViewer) refreshView() {
	if l.parsed == nil {
		l.viewport.SetContent("")
		return
	}

	lines := l.parsed.FormatColorized()
	if l.wordWrap && l.viewport.Width > 0 {
		lines = wrapLines(lines, l.viewport.Width)
	}

	if l.search != "" {
		l.matches = nil
		lower := strings.ToLower(l.search)
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), lower) {
				l.matches = append(l.matches, i)
			}
		}
	} else {
		l.matches = nil
	}

	l.viewport.SetContent(strings.Join(lines, "\n"))
}

func (l *LogViewer) Update(msg tea.Msg) (*LogViewer, tea.Cmd) {
	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	return l, cmd
}

func (l *LogViewer) View() string {
	border := styles.InactiveBorder
	if l.Active {
		border = styles.ActiveBorder
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render(l.Title)
	status := ""
	if l.search != "" && len(l.matches) > 0 {
		status = lipgloss.NewStyle().Foreground(styles.Warning).
			Render(Sprintf(" [%d/%d]", l.matchIdx+1, len(l.matches)))
	}

	header := title + status
	content := header + "\n" + l.viewport.View()

	return border.Width(l.Width - 2).Height(l.Height - 2).Render(content)
}

func wrapLines(lines []string, width int) []string {
	var result []string
	for _, line := range lines {
		if len(line) <= width {
			result = append(result, line)
			continue
		}
		for len(line) > width {
			result = append(result, line[:width])
			line = line[width:]
		}
		if len(line) > 0 {
			result = append(result, line)
		}
	}
	return result
}
