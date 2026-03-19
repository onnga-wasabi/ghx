package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/components"
)

// simulateIssuesView creates a multi-section layout (used by IssuesView)
// with n sections, each having itemCounts[i] items.
func simulateIssuesView(
	totalW, totalH int,
	nSections int,
	itemCounts []int,
	sectionIdx int,
	smartLayout bool,
	showSidebar bool,
) (rendered string, tableHeights []int) {
	listW := totalW
	sidebarW := 0
	if showSidebar {
		sidebarW = int(float64(totalW) * 0.45)
		listW = totalW - sidebarW
	}

	heights := distributeHeight(totalH, nSections, sectionIdx, smartLayout)
	tableHeights = heights

	tables := make([]*components.Table, nSections)
	for i := 0; i < nSections; i++ {
		tables[i] = components.NewTable(fmt.Sprintf("Section %d", i))
		tables[i].Width = listW
		tables[i].Height = heights[i]
		tables[i].Active = i == sectionIdx

		items := make([]components.TableItem, itemCounts[i])
		for j := range items {
			items[j] = components.TableItem{
				ID:      fmt.Sprintf("%d", j),
				Columns: []string{"✓", fmt.Sprintf("#%d", j), "Title here", "author", "💬5"},
			}
		}
		tables[i].SetItems(items)
	}

	var tableViews []string
	for _, t := range tables {
		tableViews = append(tableViews, t.View())
	}
	listPane := lipgloss.JoinVertical(lipgloss.Left, tableViews...)

	if showSidebar {
		sidebar := components.NewSidebar()
		sidebar.SetSize(sidebarW, totalH)
		sidebar.Title = "Preview"
		sidebar.SetContent("Content\nLine2\nLine3")
		rendered = lipgloss.JoinHorizontal(lipgloss.Top, listPane, sidebar.View())
	} else {
		rendered = listPane
	}

	return rendered, tableHeights
}

// simulatePRsView creates the new single-table filter layout used by PRsView.
func simulatePRsView(
	totalW, totalH int,
	nItems int,
	showSidebar bool,
) string {
	listW := totalW
	sidebarW := 0
	if showSidebar {
		sidebarW = int(float64(totalW) * 0.45)
		listW = totalW - sidebarW
	}

	filterBar := lipgloss.NewStyle().Width(listW).MaxWidth(listW).Render("Open Closed  ║  All │ Mine │ Review")
	tableH := max(4, totalH-1)

	table := components.NewTable("Pull Requests · Open · All")
	table.Width = listW
	table.Height = tableH
	table.Active = true

	items := make([]components.TableItem, nItems)
	for j := range items {
		items[j] = components.TableItem{
			ID:      fmt.Sprintf("%d", j),
			Columns: []string{"●", fmt.Sprintf("#%d", j), "Title here", "author", "3/5"},
		}
	}
	table.SetItems(items)

	listPane := lipgloss.JoinVertical(lipgloss.Left, filterBar, table.View())

	if showSidebar {
		sidebar := components.NewSidebar()
		sidebar.SetSize(sidebarW, totalH)
		sidebar.Title = "Preview"
		sidebar.SetContent("Content")
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, sidebar.View())
	}

	return listPane
}

// TestIssuesViewAllSectionsCombinations tests multi-section layout heights.
func TestIssuesViewAllSectionsCombinations(t *testing.T) {
	nSections := 3
	itemCounts := []int{5, 3, 15}
	totalH := 37
	totalW := 120

	for _, smart := range []bool{true, false} {
		for _, sidebar := range []bool{true, false} {
			for secIdx := 0; secIdx < nSections; secIdx++ {
				name := fmt.Sprintf("smart=%v sidebar=%v section=%d", smart, sidebar, secIdx)
				t.Run(name, func(t *testing.T) {
					rendered, heights := simulateIssuesView(totalW, totalH, nSections, itemCounts, secIdx, smart, sidebar)
					h := countLines(rendered)

					sum := 0
					for _, hh := range heights {
						sum += hh
					}
					if sum != totalH {
						t.Errorf("heights sum=%d want %d (heights=%v)", sum, totalH, heights)
					}
					if h != totalH {
						t.Errorf("output=%d lines, want %d (heights=%v)", h, totalH, heights)
					}
				})
			}
		}
	}
}

// TestPRsViewFilterLayout tests the new filter bar + single table layout.
func TestPRsViewFilterLayout(t *testing.T) {
	for _, totalH := range []int{20, 30, 37, 46, 50} {
		for _, sidebar := range []bool{true, false} {
			for _, nItems := range []int{0, 5, 20, 50} {
				name := fmt.Sprintf("h=%d sidebar=%v items=%d", totalH, sidebar, nItems)
				t.Run(name, func(t *testing.T) {
					rendered := simulatePRsView(120, totalH, nItems, sidebar)
					h := countLines(rendered)
					if h != totalH {
						t.Errorf("output=%d lines, want %d", h, totalH)
					}
				})
			}
		}
	}
}

// TestPRsViewScrolling tests scrolling in the single-table PRs layout.
func TestPRsViewScrolling(t *testing.T) {
	totalW := 120
	for _, totalH := range []int{25, 37, 50} {
		for _, sidebar := range []bool{true, false} {
			name := fmt.Sprintf("h=%d sidebar=%v", totalH, sidebar)
			t.Run(name, func(t *testing.T) {
				listW := totalW
				sidebarW := 0
				if sidebar {
					sidebarW = int(float64(totalW) * 0.45)
					listW = totalW - sidebarW
				}

				tableH := max(4, totalH-1)
				table := components.NewTable("Pull Requests · Open · All")
				table.Width = listW
				table.Height = tableH
				table.Active = true

				nItems := 30
				items := make([]components.TableItem, nItems)
				for j := range items {
					items[j] = components.TableItem{
						ID:      fmt.Sprintf("%d", j),
						Columns: []string{"●", fmt.Sprintf("#%d", j), "Title", "user", "3/5"},
					}
				}
				table.SetItems(items)

				renderView := func() string {
					filterBar := lipgloss.NewStyle().Width(listW).MaxWidth(listW).Render("Open ║ All")
					listPane := lipgloss.JoinVertical(lipgloss.Left, filterBar, table.View())
					if sidebar {
						sb := components.NewSidebar()
						sb.SetSize(sidebarW, totalH)
						sb.Title = "Preview"
						sb.SetContent("test")
						return lipgloss.JoinHorizontal(lipgloss.Top, listPane, sb.View())
					}
					return listPane
				}

				for step := 0; step < nItems+2; step++ {
					rendered := renderView()
					h := countLines(rendered)
					if h != totalH {
						t.Errorf("step=%d cursor=%d offset=%d → output=%d (want %d)",
							step, table.Cursor, table.Offset, h, totalH)
					}
					table.MoveDown()
				}

				table.GoToLast()
				if h := countLines(renderView()); h != totalH {
					t.Errorf("GoToLast → output=%d (want %d)", h, totalH)
				}

				table.GoToFirst()
				if h := countLines(renderView()); h != totalH {
					t.Errorf("GoToFirst → output=%d (want %d)", h, totalH)
				}
			})
		}
	}
}

// TestFullAppPipelineWithPRs tests the complete app rendering pipeline:
// tabBar + content(PRsView) + footer = terminal height
func TestFullAppPipelineWithPRs(t *testing.T) {
	for termH := 24; termH <= 50; termH++ {
		tabBarH := 2
		footerH := 1
		contentH := termH - tabBarH - footerH

		for _, sidebar := range []bool{true, false} {
			name := fmt.Sprintf("term=%d sidebar=%v", termH, sidebar)
			t.Run(name, func(t *testing.T) {
				rendered := simulatePRsView(120, contentH, 15, sidebar)
				viewH := countLines(rendered)

				content := lipgloss.NewStyle().
					Width(120).
					Height(contentH).
					MaxHeight(contentH).
					Render(rendered)

				wrappedH := countLines(content)
				if wrappedH != contentH {
					t.Errorf("view output=%d, wrapped=%d, want %d", viewH, wrappedH, contentH)
				}

				tabBar := strings.Repeat("x", 120) + "\n" + strings.Repeat("━", 120)
				footer := "help"
				output := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, footer)
				totalH := countLines(output)

				if totalH != termH {
					t.Errorf("total output=%d, want %d", totalH, termH)
				}
			})
		}
	}
}
