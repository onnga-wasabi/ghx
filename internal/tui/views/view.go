package views

import (
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type View interface {
	Name() string
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
	SetSize(width, height int)
	KeyMap() help.KeyMap
	SetFilter(query string)
}

type FilterKeyInterceptor interface {
	WantsFilterKey() bool
}

type InputModeView interface {
	IsInputMode() bool
}

// distributeHeight splits total height among n panes so that
// sum(result) == total exactly. Each pane gets at least minTableH (4)
// when total allows. With smartLayout the active pane gets the majority.
func distributeHeight(total, n, activeIdx int, smartLayout bool) []int {
	const minTableH = 4

	if n <= 0 {
		return nil
	}
	heights := make([]int, n)
	if n == 1 {
		heights[0] = total
		return heights
	}

	if smartLayout && total >= minTableH*n {
		otherH := max(minTableH, total*40/100/(n-1))
		activeH := total - otherH*(n-1)
		if activeH < minTableH {
			otherH = (total - minTableH) / (n - 1)
			activeH = total - otherH*(n-1)
		}
		for i := range heights {
			if i == activeIdx {
				heights[i] = activeH
			} else {
				heights[i] = otherH
			}
		}
		sum := 0
		for _, h := range heights {
			sum += h
		}
		heights[activeIdx] += total - sum
		return heights
	}

	baseH := total / n
	for i := range heights {
		heights[i] = baseH
	}
	heights[0] += total - baseH*n
	return heights
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	return cmd.Start()
}
