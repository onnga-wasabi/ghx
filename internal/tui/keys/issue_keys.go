package keys

import "github.com/charmbracelet/bubbles/key"

type IssueKeyMap struct {
	StateToggle key.Binding
	Close       key.Binding
	Reopen      key.Binding
	Comment     key.Binding
	Label       key.Binding
	Assign      key.Binding
}

var Issue = IssueKeyMap{
	StateToggle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "open/closed"),
	),
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "reopen"),
	),
	Comment: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "comment"),
	),
	Label: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "label"),
		key.WithDisabled(),
	),
	Assign: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "assign"),
		key.WithDisabled(),
	),
}

func (k IssueKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StateToggle, k.Close, k.Reopen, k.Comment}
}

func (k IssueKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StateToggle, k.Close, k.Reopen, k.Comment},
	}
}
