package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

	sectionExpanded []bool
	sectionCursor   int
	sectionToLine   []int // maps section index → rendered line index of header

	searching   bool
	searchInput textinput.Model
}

func NewLogViewer() *LogViewer {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.CharLimit = 256
	return &LogViewer{
		viewport:    viewport.New(0, 0),
		searchInput: ti,
	}
}

func (l *LogViewer) SetContent(raw string) {
	l.content = raw
	l.parsed = ParseLogs(raw)
	if l.parsed != nil {
		l.sectionExpanded = make([]bool, len(l.parsed.Steps))
		for i := range l.sectionExpanded {
			l.sectionExpanded[i] = true
		}
	} else {
		l.sectionExpanded = nil
	}
	l.sectionCursor = 0
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

func (l *LogViewer) IsSearching() bool {
	return l.searching
}

func (l *LogViewer) StartSearch() tea.Cmd {
	l.searching = true
	l.searchInput.SetValue(l.search)
	return l.searchInput.Focus()
}

func (l *LogViewer) submitSearch() {
	l.search = l.searchInput.Value()
	l.matchIdx = 0
	l.searching = false
	l.searchInput.Blur()
	l.refreshView()
}

func (l *LogViewer) cancelSearch() {
	l.searching = false
	l.searchInput.Blur()
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

func (l *LogViewer) ToggleCurrentSection() {
	if l.sectionCursor < 0 || l.sectionCursor >= len(l.sectionExpanded) {
		return
	}
	l.sectionExpanded[l.sectionCursor] = !l.sectionExpanded[l.sectionCursor]
	l.refreshView()
	l.scrollToCurrentSection()
}

func (l *LogViewer) ExpandAll() {
	for i := range l.sectionExpanded {
		l.sectionExpanded[i] = true
	}
	l.refreshView()
	l.scrollToCurrentSection()
}

func (l *LogViewer) CollapseAll() {
	for i := range l.sectionExpanded {
		l.sectionExpanded[i] = false
	}
	l.refreshView()
	l.scrollToCurrentSection()
}

func (l *LogViewer) NextSection() {
	if len(l.sectionExpanded) == 0 {
		return
	}
	if l.sectionCursor < len(l.sectionExpanded)-1 {
		l.sectionCursor++
	}
	l.refreshView()
	l.scrollToCurrentSection()
}

func (l *LogViewer) PrevSection() {
	if len(l.sectionExpanded) == 0 {
		return
	}
	if l.sectionCursor > 0 {
		l.sectionCursor--
	}
	l.refreshView()
	l.scrollToCurrentSection()
}

func (l *LogViewer) scrollToCurrentSection() {
	if l.sectionCursor < 0 || l.sectionCursor >= len(l.sectionToLine) {
		return
	}
	headerLine := l.sectionToLine[l.sectionCursor]
	vpH := l.viewport.Height
	yOff := l.viewport.YOffset

	if headerLine < yOff {
		l.viewport.SetYOffset(headerLine)
	} else if headerLine >= yOff+vpH {
		l.viewport.SetYOffset(headerLine - vpH + 1)
	}
}

func (l *LogViewer) refreshView() {
	if l.parsed == nil || len(l.parsed.Steps) == 0 {
		l.viewport.SetContent(l.content)
		l.sectionToLine = nil
		return
	}

	if l.sectionCursor >= len(l.parsed.Steps) {
		l.sectionCursor = max(0, len(l.parsed.Steps)-1)
	}

	errorStyle := lipgloss.NewStyle().Foreground(styles.Error).Bold(true)
	warnStyle := lipgloss.NewStyle().Foreground(styles.Warning)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	countStyle := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true)

	stepStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	collapsedStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	cursorStepStyle := lipgloss.NewStyle().Foreground(styles.Warning).Bold(true)
	cursorCollapsedStyle := lipgloss.NewStyle().Foreground(styles.Warning).Bold(true)

	var lines []string
	l.sectionToLine = make([]int, len(l.parsed.Steps))

	for i, step := range l.parsed.Steps {
		expanded := i < len(l.sectionExpanded) && l.sectionExpanded[i]
		isCursor := i == l.sectionCursor
		l.sectionToLine[i] = len(lines)

		if expanded {
			sStyle := stepStyle
			prefix := "  ▾ "
			if isCursor {
				sStyle = cursorStepStyle
				prefix = "▸ ▾ "
			}
			lines = append(lines, sStyle.Render(prefix+step.Name))

			for _, ll := range step.Lines {
				switch ll.Level {
				case "error":
					lines = append(lines, errorStyle.Render("    ✗ "+ll.Text))
				case "warning":
					lines = append(lines, warnStyle.Render("    ⚠ "+ll.Text))
				case "debug":
					lines = append(lines, mutedStyle.Render("    "+ll.Text))
				default:
					lines = append(lines, "    "+ll.Text)
				}
			}
		} else {
			cStyle := collapsedStyle
			prefix := "  ▸ "
			if isCursor {
				cStyle = cursorCollapsedStyle
				prefix = "▸ ▸ "
			}
			header := cStyle.Render(prefix+step.Name) +
				" " + countStyle.Render(fmt.Sprintf("(%d lines)", len(step.Lines)))
			lines = append(lines, header)
		}
	}

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
	if l.searching {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "enter":
				l.submitSearch()
				return l, nil
			case "esc":
				l.cancelSearch()
				return l, nil
			}
			var cmd tea.Cmd
			l.searchInput, cmd = l.searchInput.Update(msg)
			return l, cmd
		}
	}
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

	expanded := 0
	total := len(l.sectionExpanded)
	for _, e := range l.sectionExpanded {
		if e {
			expanded++
		}
	}
	sectionInfo := ""
	if total > 0 {
		sectionInfo = lipgloss.NewStyle().Foreground(styles.Muted).
			Render(fmt.Sprintf(" [%d/%d]", expanded, total))
	}

	header := title + status + sectionInfo

	var content string
	if l.searching {
		content = header + "\n" + l.searchInput.View() + "\n" + l.viewport.View()
	} else {
		content = header + "\n" + l.viewport.View()
	}

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
