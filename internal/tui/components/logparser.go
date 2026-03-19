package components

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

var (
	timestampRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z\s*`)
	groupRe     = regexp.MustCompile(`##\[group\](.*)`)
	endGroupRe  = regexp.MustCompile(`##\[endgroup\]`)
	commandRe   = regexp.MustCompile(`##\[(error|warning|notice|debug)\](.*)`)
)

type ParsedLogs struct {
	Steps []StepLog
	Raw   []string
}

type StepLog struct {
	Name  string
	Lines []LogLine
}

type LogLine struct {
	Text  string
	Level string // "error", "warning", "notice", "debug", ""
}

func Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func ParseLogs(raw string) *ParsedLogs {
	lines := strings.Split(raw, "\n")
	parsed := &ParsedLogs{Raw: lines}

	var currentStep *StepLog
	defaultStep := StepLog{Name: "Output"}

	for _, line := range lines {
		line = timestampRe.ReplaceAllString(line, "")

		if m := groupRe.FindStringSubmatch(line); m != nil {
			if currentStep != nil {
				parsed.Steps = append(parsed.Steps, *currentStep)
			}
			currentStep = &StepLog{Name: m[1]}
			continue
		}
		if endGroupRe.MatchString(line) {
			if currentStep != nil {
				parsed.Steps = append(parsed.Steps, *currentStep)
				currentStep = nil
			}
			continue
		}

		ll := LogLine{Text: line}
		if m := commandRe.FindStringSubmatch(line); m != nil {
			ll.Level = m[1]
			ll.Text = m[2]
		}

		if currentStep != nil {
			currentStep.Lines = append(currentStep.Lines, ll)
		} else {
			defaultStep.Lines = append(defaultStep.Lines, ll)
		}
	}

	if currentStep != nil {
		parsed.Steps = append(parsed.Steps, *currentStep)
	}
	if len(defaultStep.Lines) > 0 && len(parsed.Steps) == 0 {
		parsed.Steps = append(parsed.Steps, defaultStep)
	}

	return parsed
}

func (p *ParsedLogs) FormatColorized() []string {
	var lines []string

	errorStyle := lipgloss.NewStyle().Foreground(styles.Error).Bold(true)
	warnStyle := lipgloss.NewStyle().Foreground(styles.Warning)
	stepStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.Muted)

	for _, step := range p.Steps {
		lines = append(lines, stepStyle.Render("▸ "+step.Name))
		for _, ll := range step.Lines {
			switch ll.Level {
			case "error":
				lines = append(lines, errorStyle.Render("  ✗ "+ll.Text))
			case "warning":
				lines = append(lines, warnStyle.Render("  ⚠ "+ll.Text))
			case "debug":
				lines = append(lines, mutedStyle.Render("  "+ll.Text))
			default:
				lines = append(lines, "  "+ll.Text)
			}
		}
		lines = append(lines, "")
	}
	return lines
}
