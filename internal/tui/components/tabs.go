package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type Tabs struct {
	Items  []string
	Active int
	Width  int
}

func NewTabs(items []string) *Tabs {
	return &Tabs{Items: items}
}

func (t *Tabs) Next() {
	t.Active = (t.Active + 1) % len(t.Items)
}

func (t *Tabs) Prev() {
	t.Active = (t.Active - 1 + len(t.Items)) % len(t.Items)
}

func (t *Tabs) SetActive(idx int) {
	if idx >= 0 && idx < len(t.Items) {
		t.Active = idx
	}
}

func (t *Tabs) View() string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1a1b26")).
		Background(styles.Primary).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 2)

	numStyle := lipgloss.NewStyle().
		Foreground(styles.Muted)

	var tabs []string
	for i, item := range t.Items {
		label := fmt.Sprintf("%d:%s", i+1, item)
		if i == t.Active {
			tabs = append(tabs, activeStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveStyle.Render(numStyle.Render(fmt.Sprintf("%d:", i+1))+item))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	pad := max(0, t.Width-lipgloss.Width(row))
	bg := lipgloss.NewStyle().
		Background(lipgloss.Color("#24283b")).
		Width(pad).
		Render("")
	row += bg

	border := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Render(strings.Repeat("━", max(0, t.Width)))

	return row + "\n" + border
}
