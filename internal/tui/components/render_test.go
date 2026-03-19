package components

import (
	"fmt"
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

func maxLineWidth(s string) int {
	w := 0
	for _, line := range strings.Split(s, "\n") {
		lw := lipgloss.Width(line)
		if lw > w {
			w = lw
		}
	}
	return w
}

func TestLipglossBorderHeightBehavior(t *testing.T) {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555"))

	tests := []struct {
		name        string
		w, h        int
		contentH    int
		wantOutputH int
	}{
		{"W=20 H=10 content=8", 20, 10, 8, -1},
		{"W=20 H=10 content=6", 20, 10, 6, -1},
		{"W=20 H=8 content=8", 20, 8, 8, -1},
		{"W=20 H=8 content=6", 20, 8, 6, -1},
		{"W=20 H=4 content=4", 20, 4, 4, -1},
		{"W=20 H=4 content=2", 20, 4, 2, -1},
		{"W=20 H=2 content=2", 20, 2, 2, -1},
		{"W=20 H=2 content=0", 20, 2, 0, -1},
		{"W=20 H=0 content=0", 20, 0, 0, -1},
		{"W=20 H=-1 content=0", 20, -1, 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lines []string
			for i := 0; i < tt.contentH; i++ {
				lines = append(lines, "x")
			}
			content := strings.Join(lines, "\n")

			rendered := border.Width(tt.w).Height(tt.h).Render(content)
			actualH := countLines(rendered)
			actualW := maxLineWidth(rendered)

			t.Logf("Width(%d).Height(%d) content=%d lines → output: %d lines, %d wide",
				tt.w, tt.h, tt.contentH, actualH, actualW)
			t.Logf("Rendered:\n%s", rendered)
		})
	}
}

// Verifies: border.Width(W).Height(H) → output = H+2 lines, W+2 wide.
// Height/Width in lipgloss v1.1.0 specify INNER dimensions; border adds 2.
func TestLipglossBorderExactDimension(t *testing.T) {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555"))

	for innerH := 1; innerH <= 15; innerH++ {
		for contentH := 0; contentH <= innerH; contentH++ {
			var lines []string
			for i := 0; i < contentH; i++ {
				lines = append(lines, strings.Repeat("x", 10))
			}
			content := strings.Join(lines, "\n")

			rendered := border.Width(20).Height(innerH).Render(content)
			actualH := countLines(rendered)
			wantH := innerH + 2

			if actualH != wantH {
				t.Errorf("border.Height(%d) content=%d → output=%d (want %d = innerH+2)",
					innerH, contentH, actualH, wantH)
			}
		}
	}
}

func TestTableViewExactHeight(t *testing.T) {
	for h := 3; h <= 30; h++ {
		t.Run("", func(t *testing.T) {
			tbl := NewTable("Test Section")
			tbl.Width = 40
			tbl.Height = h
			tbl.Active = true

			items := make([]TableItem, 20)
			for i := range items {
				items[i] = TableItem{ID: "x", Columns: []string{"col1", "col2"}}
			}
			tbl.SetItems(items)

			rendered := tbl.View()
			actualH := countLines(rendered)

			if actualH != h {
				t.Errorf("Table Height=%d items=20 → output=%d lines (want %d)",
					h, actualH, h)
			}
		})
	}
}

func TestTableViewExactHeightEmpty(t *testing.T) {
	for h := 3; h <= 30; h++ {
		t.Run("", func(t *testing.T) {
			tbl := NewTable("Empty Table")
			tbl.Width = 40
			tbl.Height = h
			tbl.Active = false
			tbl.Placeholder = "No items"

			rendered := tbl.View()
			actualH := countLines(rendered)

			if actualH != h {
				t.Errorf("Empty Table Height=%d → output=%d lines (want %d)",
					h, actualH, h)
			}
		})
	}
}

func TestTableViewScrollConsistency(t *testing.T) {
	for h := 4; h <= 20; h++ {
		tbl := NewTable("Scroll Test")
		tbl.Width = 40
		tbl.Height = h
		tbl.Active = true

		items := make([]TableItem, 50)
		for i := range items {
			items[i] = TableItem{ID: "x", Columns: []string{"item"}}
		}
		tbl.SetItems(items)

		for step := 0; step < 55; step++ {
			rendered := tbl.View()
			actualH := countLines(rendered)
			if actualH != h {
				t.Errorf("Height=%d step=%d cursor=%d offset=%d → output=%d lines (want %d)",
					h, step, tbl.Cursor, tbl.Offset, actualH, h)
			}
			tbl.MoveDown()
		}

		for step := 0; step < 55; step++ {
			rendered := tbl.View()
			actualH := countLines(rendered)
			if actualH != h {
				t.Errorf("Height=%d step-up=%d cursor=%d offset=%d → output=%d lines (want %d)",
					h, step, tbl.Cursor, tbl.Offset, actualH, h)
			}
			tbl.MoveUp()
		}

		tbl.GoToLast()
		rendered := tbl.View()
		if countLines(rendered) != h {
			t.Errorf("Height=%d GoToLast → output=%d lines", h, countLines(rendered))
		}

		tbl.GoToFirst()
		rendered = tbl.View()
		if countLines(rendered) != h {
			t.Errorf("Height=%d GoToFirst → output=%d lines", h, countLines(rendered))
		}
	}
}

func TestSidebarViewExactHeight(t *testing.T) {
	for h := 5; h <= 30; h++ {
		t.Run("", func(t *testing.T) {
			sb := NewSidebar()
			sb.SetSize(30, h)
			sb.Title = "Preview"
			sb.SetContent("Some content\nLine 2\nLine 3\nLine 4\nLine 5")

			rendered := sb.View()
			actualH := countLines(rendered)

			if actualH != h {
				t.Errorf("Sidebar Height=%d → output=%d lines (want %d)",
					h, actualH, h)
			}
		})
	}
}

func TestLogViewerExactHeight(t *testing.T) {
	for h := 4; h <= 30; h++ {
		t.Run("", func(t *testing.T) {
			lv := NewLogViewer()
			lv.SetSize(40, h)
			lv.Title = "Logs"
			lv.Active = true
			lv.SetContent("line1\nline2\nline3\nline4\nline5")

			rendered := lv.View()
			actualH := countLines(rendered)

			if actualH != h {
				t.Errorf("LogViewer Height=%d → output=%d lines (want %d)",
					h, actualH, h)
			}
		})
	}
}

func TestTableViewWithMultibyteContent(t *testing.T) {
	for _, h := range []int{4, 6, 10, 20, 30} {
		t.Run(fmt.Sprintf("h=%d", h), func(t *testing.T) {
			tbl := NewTable("セクション")
			tbl.Width = 60
			tbl.Height = h
			tbl.Active = true

			items := []TableItem{
				{ID: "1", Columns: []string{"✓", "#96", "staging deployの通信失敗を解消", "user", "9/9"}},
				{ID: "2", Columns: []string{"✗", "#95", "staging デプロイ変化: l2-edge🔒🔒", "user", "9/9"}},
				{ID: "3", Columns: []string{"⏳", "#94", "l2-edge LBスキーム不整合を解消…", "user", "9/9"}},
				{ID: "4", Columns: []string{"✓", "#93", "backend ヘルスチェック修正 長いタイトルのテスト用テキスト", "user", "9/9"}},
				{ID: "5", Columns: []string{"●", "#92", "💬 コメント付きPR", "user", "💬5"}},
				{ID: "6", Columns: []string{"✓", "#91", "普通のASCII title", "user", "9/9"}},
				{ID: "7", Columns: []string{"✓", "#90", "混合 mixed テキスト with 日本語 and English", "user", "9/9"}},
				{ID: "8", Columns: []string{"✓", "#89", "infra apply失敗を正しく💀💀💀修正", "user", "9/9"}},
				{ID: "9", Columns: []string{"✓", "#88", "全角スペース　を含むタイトル", "user", "9/9"}},
				{ID: "10", Columns: []string{"✓", "#87", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "user", "9/9"}},
			}
			tbl.SetItems(items)

			for step := 0; step < len(items)+2; step++ {
				rendered := tbl.View()
				actualH := countLines(rendered)
				if actualH != h {
					t.Errorf("step=%d cursor=%d offset=%d → output=%d lines (want %d)",
						step, tbl.Cursor, tbl.Offset, actualH, h)
				}
				tbl.MoveDown()
			}

			tbl.GoToLast()
			rendered := tbl.View()
			if countLines(rendered) != h {
				t.Errorf("GoToLast → output=%d lines (want %d)", countLines(rendered), h)
			}
		})
	}
}

func TestJoinVerticalTablesSumHeight(t *testing.T) {
	heights := []int{4, 4, 15, 4}
	totalH := 0
	for _, h := range heights {
		totalH += h
	}

	var views []string
	for i, h := range heights {
		tbl := NewTable("Section " + string(rune('A'+i)))
		tbl.Width = 40
		tbl.Height = h
		tbl.Active = i == 2

		items := make([]TableItem, 10)
		for j := range items {
			items[j] = TableItem{ID: "x", Columns: []string{"item"}}
		}
		tbl.SetItems(items)
		views = append(views, tbl.View())
	}

	joined := lipgloss.JoinVertical(lipgloss.Left, views...)
	actualH := countLines(joined)

	if actualH != totalH {
		t.Errorf("JoinVertical of tables %v → output=%d lines (want %d)",
			heights, actualH, totalH)
	}
}
