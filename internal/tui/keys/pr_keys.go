package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	StateToggle key.Binding
	Enhance     key.Binding
	Diff        key.Binding
	Approve     key.Binding
	Merge       key.Binding
	Checkout    key.Binding
	Close       key.Binding
	Ready       key.Binding
	Comment     key.Binding
}

var PR = PRKeyMap{
	StateToggle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "open/closed"),
	),
	Enhance: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "CI checks"),
	),
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
	Approve: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "approve"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "merge"),
	),
	Checkout: key.NewBinding(
		key.WithKeys("ctrl+o"),
		key.WithHelp("C-o", "checkout"),
		key.WithDisabled(),
	),
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
	Ready: key.NewBinding(
		key.WithKeys("W"),
		key.WithHelp("W", "mark ready"),
		key.WithDisabled(),
	),
	Comment: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "comment"),
	),
}

func (k PRKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StateToggle, k.Enhance, k.Diff, k.Approve, k.Merge}
}

func (k PRKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StateToggle, k.Enhance, k.Diff},
		{k.Approve, k.Merge, k.Close, k.Comment},
	}
}
