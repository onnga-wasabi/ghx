package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type Filter struct {
	input   textinput.Model
	active  bool
	value   string
	width   int
}

func NewFilter() *Filter {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.CharLimit = 256
	return &Filter{input: ti}
}

func (f *Filter) IsActive() bool {
	return f.active
}

func (f *Filter) Value() string {
	return f.value
}

func (f *Filter) SetWidth(w int) {
	f.width = w
	f.input.Width = w - 4
}

func (f *Filter) Activate() tea.Cmd {
	f.active = true
	f.input.SetValue(f.value)
	return f.input.Focus()
}

func (f *Filter) Deactivate() {
	f.active = false
	f.input.Blur()
}

func (f *Filter) Submit() {
	f.value = f.input.Value()
	f.Deactivate()
}

func (f *Filter) Cancel() {
	f.input.SetValue(f.value)
	f.Deactivate()
}

func (f *Filter) Clear() {
	f.value = ""
	f.input.SetValue("")
	f.Deactivate()
}

func (f *Filter) Update(msg tea.Msg) (*Filter, tea.Cmd) {
	if !f.active {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			f.Submit()
			return f, nil
		case "esc":
			f.Cancel()
			return f, nil
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return f, cmd
}

func (f *Filter) View() string {
	if f.active {
		return f.input.View()
	}
	if f.value != "" {
		return lipgloss.NewStyle().Foreground(styles.Muted).Render("filter: ") +
			lipgloss.NewStyle().Foreground(styles.Warning).Render(f.value)
	}
	return ""
}
