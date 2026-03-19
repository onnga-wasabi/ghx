package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type TableItem struct {
	ID       string
	Columns  []string
	Selected bool
}

type Table struct {
	Items       []TableItem
	Cursor      int
	Offset      int
	Width       int
	Height      int
	Title       string
	Active      bool
	ColWidth    []int
	Placeholder string
}

func NewTable(title string) *Table {
	return &Table{Title: title}
}

func (t *Table) SetItems(items []TableItem) {
	t.Items = items
	if t.Cursor >= len(items) {
		t.Cursor = max(0, len(items)-1)
	}
	t.Placeholder = ""
}

func (t *Table) ClearItems(placeholder string) {
	t.Items = nil
	t.Cursor = 0
	t.Offset = 0
	t.Placeholder = placeholder
}

func (t *Table) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
		if t.Cursor < t.Offset {
			t.Offset = t.Cursor
		}
	}
}

func (t *Table) MoveDown() {
	if t.Cursor < len(t.Items)-1 {
		t.Cursor++
		if vh := t.visibleHeight(); vh > 0 && t.Cursor >= t.Offset+vh {
			t.Offset = t.Cursor - vh + 1
		}
	}
}

func (t *Table) GoToFirst() {
	t.Cursor = 0
	t.Offset = 0
}

func (t *Table) GoToLast() {
	if len(t.Items) > 0 {
		t.Cursor = len(t.Items) - 1
		if vh := t.visibleHeight(); vh > 0 && t.Cursor >= vh {
			t.Offset = t.Cursor - vh + 1
		}
	}
}

func (t *Table) SelectedItem() *TableItem {
	if t.Cursor >= 0 && t.Cursor < len(t.Items) {
		return &t.Items[t.Cursor]
	}
	return nil
}

// visibleHeight returns the number of data rows that fit in the table.
// The rendered table is: border(1) + title(1) + rows(vh) + padding + border(1).
func (t *Table) visibleHeight() int {
	return max(0, t.Height-3)
}

func (t *Table) ensureCursorVisible() {
	vh := t.visibleHeight()
	if vh <= 0 {
		return
	}
	if t.Cursor < t.Offset {
		t.Offset = t.Cursor
	}
	if t.Cursor >= t.Offset+vh {
		t.Offset = t.Cursor - vh + 1
	}
}

// View renders the table to exactly t.Height lines.
func (t *Table) View() string {
	if t.Height <= 0 {
		return ""
	}

	t.ensureCursorVisible()

	border := styles.InactiveBorder
	if t.Active {
		border = styles.ActiveBorder
	}

	innerH := max(0, t.Height-2)

	var lines []string

	if innerH > 0 {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
		lines = append(lines, titleStyle.Render(t.Title))
	}

	rowsAvail := innerH - len(lines)

	if rowsAvail > 0 {
		if len(t.Items) == 0 && t.Placeholder != "" {
			ph := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render(t.Placeholder)
			lines = append(lines, ph)
		} else {
			end := min(t.Offset+rowsAvail, len(t.Items))
			for i := t.Offset; i < end; i++ {
				lines = append(lines, t.renderRow(i))
			}
		}
	}

	for len(lines) < innerH {
		lines = append(lines, strings.Repeat(" ", max(0, t.Width-4)))
	}
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	content := strings.Join(lines, "\n")
	return border.Width(t.Width - 2).Height(innerH).Render(content)
}

func (t *Table) renderRow(idx int) string {
	item := t.Items[idx]
	cols := strings.Join(item.Columns, " ")

	maxW := max(0, t.Width-6)
	if ansi.StringWidth(cols) > maxW {
		cols = ansi.Truncate(cols, maxW-1, "…")
	}

	if idx == t.Cursor {
		style := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
		return style.Render("▸ " + cols)
	}

	return "  " + lipgloss.NewStyle().Foreground(styles.Text).Render(cols)
}
