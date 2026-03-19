package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type Help struct {
	model   help.Model
	showAll bool
	width   int
}

func NewHelp() *Help {
	h := help.New()
	h.ShortSeparator = " │ "
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(styles.Muted)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(styles.Text)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(styles.Muted)
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(styles.Muted)
	return &Help{model: h}
}

func (h *Help) SetWidth(w int) {
	h.width = w
	h.model.Width = w
}

func (h *Help) Toggle() {
	h.showAll = !h.showAll
}

func (h *Help) IsShowingAll() bool {
	return h.showAll
}

// View always returns the short (one-line) help for the footer.
func (h *Help) View(global help.KeyMap, contextual ...help.KeyMap) string {
	var allBindings []key.Binding
	allBindings = append(allBindings, global.ShortHelp()...)
	for _, km := range contextual {
		allBindings = append(allBindings, km.ShortHelp()...)
	}
	short := &shortKeyMap{bindings: allBindings}
	return h.model.ShortHelpView(short.ShortHelp())
}

// FullView returns the full keybindings view for the floating overlay.
func (h *Help) FullView(global help.KeyMap, contextual ...help.KeyMap) string {
	var combined [][]key.Binding
	combined = append(combined, global.FullHelp()...)
	for _, km := range contextual {
		combined = append(combined, km.FullHelp()...)
	}
	full := &fullKeyMap{groups: combined}
	return h.model.FullHelpView(full.FullHelp())
}

type shortKeyMap struct{ bindings []key.Binding }

func (s *shortKeyMap) ShortHelp() []key.Binding  { return s.bindings }
func (s *shortKeyMap) FullHelp() [][]key.Binding  { return nil }

type fullKeyMap struct{ groups [][]key.Binding }

func (f *fullKeyMap) ShortHelp() []key.Binding  { return nil }
func (f *fullKeyMap) FullHelp() [][]key.Binding  { return f.groups }
