package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func TestDistributeHeightSumsCorrectly(t *testing.T) {
	testCases := []struct {
		total, n, active int
		smart            bool
	}{
		{37, 4, 0, true},
		{37, 4, 1, true},
		{37, 4, 2, true},
		{37, 4, 3, true},
		{37, 4, 0, false},
		{27, 4, 0, true},
		{27, 4, 2, true},
		{20, 4, 0, true},
		{16, 4, 0, true},
		{15, 4, 0, true},
		{12, 3, 1, true},
		{40, 2, 0, true},
		{40, 1, 0, true},
		{10, 1, 0, false},
	}

	for _, tc := range testCases {
		heights := distributeHeight(tc.total, tc.n, tc.active, tc.smart)
		sum := 0
		for _, h := range heights {
			sum += h
		}
		if sum != tc.total {
			t.Errorf("distributeHeight(%d, %d, %d, %v) = %v (sum=%d, want %d)",
				tc.total, tc.n, tc.active, tc.smart, heights, sum, tc.total)
		}

		for i, h := range heights {
			if h < 0 {
				t.Errorf("distributeHeight(%d, %d, %d, %v)[%d] = %d (negative!)",
					tc.total, tc.n, tc.active, tc.smart, i, h)
			}
		}
	}
}

func TestDistributeHeightMinimum(t *testing.T) {
	heights := distributeHeight(37, 4, 0, true)
	for i, h := range heights {
		if h < 4 {
			t.Errorf("heights[%d] = %d, want >= 4", i, h)
		}
	}
}

func TestTabBarHeight(t *testing.T) {
	// Simulate tabs rendering
	tabNames := []string{"PRs", "Issues", "Actions", "Notifications"}
	var tabs []string
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 2)
		tabs = append(tabs, style.Render(name))
		_ = i
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	border := strings.Repeat("━", 80)
	tabBar := row + "\n" + border

	h := countLines(tabBar)
	if h != 2 {
		t.Errorf("tabBar height = %d, want 2", h)
	}
}

func TestContentWrapperExactHeight(t *testing.T) {
	for contentH := 0; contentH <= 40; contentH++ {
		targetH := 30

		var lines []string
		for i := 0; i < contentH; i++ {
			lines = append(lines, strings.Repeat("x", 50))
		}
		rawContent := strings.Join(lines, "\n")

		content := lipgloss.NewStyle().
			Width(80).
			Height(targetH).
			MaxHeight(targetH).
			Render(rawContent)

		h := countLines(content)
		if h != targetH {
			t.Errorf("content wrapper: content=%d lines, target=%d → output=%d (want %d)",
				contentH, targetH, h, targetH)
		}
	}
}

func TestFullAppOutputHeight(t *testing.T) {
	for termH := 20; termH <= 50; termH++ {
		tabBarHeight := 2
		footerH := 1
		contentHeight := termH - tabBarHeight - footerH

		tabBar := strings.Repeat("x", 80) + "\n" + strings.Repeat("━", 80)
		if countLines(tabBar) != tabBarHeight {
			t.Fatalf("tabBar lines = %d", countLines(tabBar))
		}

		var contentLines []string
		for i := 0; i < contentHeight; i++ {
			contentLines = append(contentLines, strings.Repeat(".", 80))
		}
		rawContent := strings.Join(contentLines, "\n")

		content := lipgloss.NewStyle().
			Width(80).
			Height(contentHeight).
			MaxHeight(contentHeight).
			Render(rawContent)

		footer := "help text here"
		if countLines(footer) != footerH {
			t.Fatalf("footer lines = %d", countLines(footer))
		}

		output := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, footer)
		h := countLines(output)

		if h != termH {
			t.Errorf("termH=%d → output=%d lines (tabBar=%d content=%d footer=%d)",
				termH, h, countLines(tabBar), countLines(content), countLines(footer))
		}
	}
}

func TestFullAppOutputHeightWithExpandedHelp(t *testing.T) {
	termH := 40
	tabBarHeight := 2

	footerLines := []string{"key1: desc1", "key2: desc2", "key3: desc3"}
	footer := strings.Join(footerLines, "\n")
	footerH := countLines(footer)

	contentHeight := termH - tabBarHeight - footerH

	tabBar := strings.Repeat("x", 80) + "\n" + strings.Repeat("━", 80)

	var contentLns []string
	for i := 0; i < contentHeight; i++ {
		contentLns = append(contentLns, strings.Repeat(".", 80))
	}
	rawContent := strings.Join(contentLns, "\n")

	content := lipgloss.NewStyle().
		Width(80).
		Height(contentHeight).
		MaxHeight(contentHeight).
		Render(rawContent)

	output := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, footer)
	h := countLines(output)

	if h != termH {
		t.Errorf("termH=%d footerH=%d contentH=%d → output=%d lines (want %d)",
			termH, footerH, contentHeight, h, termH)
	}
}
