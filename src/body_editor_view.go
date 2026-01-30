package src

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BodyEditorViewModel struct {
	bodyFields    []BodyField
	cursor        int
	viewportStart int
	maxVisible    int
	selected      *string
	quitting      bool
	styles        *Styles
}

type BodyField struct {
	Key   string
	Value string
}

func NewBodyEditorViewModel(body interface{}) BodyEditorViewModel {
	var fields []BodyField

	// Convert body to map
	if body != nil {
		bodyMap, ok := body.(map[string]interface{})
		if ok {
			for key, value := range bodyMap {
				// Convert value to string
				var valueStr string
				switch v := value.(type) {
				case string:
					valueStr = v
				case float64:
					valueStr = fmt.Sprintf("%.0f", v)
				case bool:
					valueStr = fmt.Sprintf("%t", v)
				default:
					jsonBytes, _ := json.Marshal(v)
					valueStr = string(jsonBytes)
				}

				fields = append(fields, BodyField{
					Key:   key,
					Value: valueStr,
				})
			}
		}
	}

	return BodyEditorViewModel{
		bodyFields:    fields,
		cursor:        0,
		viewportStart: 0,
		maxVisible:    10,
		quitting:      false,
		styles:        DefaultStyles(),
	}
}

func (m BodyEditorViewModel) Init() tea.Cmd {
	return nil
}

func (m BodyEditorViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			*m.selected = ExitSignal
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewportStart+2 && m.viewportStart > 0 {
					m.viewportStart--
				}
			} else {
				m.cursor = len(m.bodyFields) - 1
				if len(m.bodyFields) > m.maxVisible {
					m.viewportStart = len(m.bodyFields) - m.maxVisible
				} else {
					m.viewportStart = 0
				}
			}

		case "down", "j":
			if m.cursor < len(m.bodyFields)-1 {
				m.cursor++
				if m.cursor >= m.viewportStart+m.maxVisible-2 {
					m.viewportStart++
				}
			} else {
				m.cursor = 0
				m.viewportStart = 0
			}

		case "enter":
			if len(m.bodyFields) > 0 && m.cursor < len(m.bodyFields) {
				selectedField := m.bodyFields[m.cursor]
				*m.selected = fmt.Sprintf("%s|%s", selectedField.Key, selectedField.Value)
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m BodyEditorViewModel) View() string {
	if m.quitting {
		return ""
	}

	var b string

	b += "\n"
	b += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", m.styles.TitleColor)
	b += m.styles.Text("  Edit Body Fields\n", m.styles.SelectedTitleColor)
	b += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", m.styles.TitleColor)
	b += "\n"

	if len(m.bodyFields) == 0 {
		b += m.styles.FooterStyle.Render("  No body fields to edit\n")
	} else {
		visibleEnd := m.viewportStart + m.maxVisible
		if visibleEnd > len(m.bodyFields) {
			visibleEnd = len(m.bodyFields)
		}

		if m.viewportStart > 0 {
			b += m.styles.FooterStyle.Render("  ⬆ More fields above...\n")
			b += "\n"
		}

		for i := m.viewportStart; i < visibleEnd; i++ {
			field := m.bodyFields[i]

			var itemBox lipgloss.Style
			if m.cursor == i {
				itemBox = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(m.styles.SelectedTitleColor).
					Padding(0, 1).
					Width(70).
					MarginLeft(2)
			} else {
				itemBox = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(m.styles.MutedBorderColor).
					Padding(0, 1).
					Width(70)
			}

			titleStyle := lipgloss.NewStyle().Bold(true)
			var valueColor lipgloss.Color
			if m.cursor == i {
				titleStyle = titleStyle.Foreground(m.styles.SelectedTitleColor)
				valueColor = m.styles.FooterColor
			} else {
				titleStyle = titleStyle.Foreground(m.styles.MutedTitleColor)
				valueColor = m.styles.MutedTitleColor
			}

			valueStyle := lipgloss.NewStyle().
				Foreground(valueColor).
				Width(66).
				Italic(true)

			content := fmt.Sprintf("%s\n%s",
				titleStyle.Render(field.Key),
				valueStyle.Render(field.Value),
			)

			b += itemBox.Render(content)
			b += "\n"
		}

		if visibleEnd < len(m.bodyFields) {
			b += "\n"
			b += m.styles.FooterStyle.Render("  ⬇ More fields below...\n")
		}
	}

	b += "\n"
	b += m.styles.FooterStyle.Render("  ↑↓/jk navigate • enter edit • q/esc back\n")

	return b
}

func BodyEditorView(body interface{}, selected *string) {
	m := NewBodyEditorViewModel(body)
	m.selected = selected

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("BodyEditorView -> ", err)
		os.Exit(1)
	}
}
