package src

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BodyEditorViewModel struct {
	bodyFields    []BodyField
	cursor        int
	viewportStart int
	maxVisible    int
	editMode      bool
	textInput     textinput.Model
	editingIndex  int
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

	ti := textinput.New()
	ti.CharLimit = 500
	ti.Width = 60

	return BodyEditorViewModel{
		bodyFields:    fields,
		cursor:        0,
		viewportStart: 0,
		maxVisible:    10,
		editMode:      false,
		textInput:     ti,
		editingIndex:  -1,
		quitting:      false,
		styles:        DefaultStyles(),
	}
}

func (m BodyEditorViewModel) Init() tea.Cmd {
	return nil
}

func (m BodyEditorViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If in edit mode, handle text input
	if m.editMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				// Cancel edit
				m.editMode = false
				m.editingIndex = -1
				return m, nil
			case "enter":
				// Save edit
				if m.editingIndex >= 0 && m.editingIndex < len(m.bodyFields) {
					m.bodyFields[m.editingIndex].Value = strings.TrimSpace(m.textInput.Value())
				}
				m.editMode = false
				m.editingIndex = -1
				return m, nil
			}
		}

		// Update text input
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Normal navigation mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			*m.selected = ExitSignal
			m.quitting = true
			return m, tea.Quit

		case "esc":
			// Save and exit
			result := ""
			for i, field := range m.bodyFields {
				if i > 0 {
					result += "||"
				}
				result += field.Key + "|" + field.Value
			}
			*m.selected = result
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewportStart {
					m.viewportStart--
				}
			}

		case "down", "j":
			if m.cursor < len(m.bodyFields)-1 {
				m.cursor++
				if m.cursor >= m.viewportStart+m.maxVisible {
					m.viewportStart++
				}
			}

		case "enter", "e":
			// Enter edit mode for current field
			if m.cursor < len(m.bodyFields) {
				m.editMode = true
				m.editingIndex = m.cursor
				m.textInput.SetValue(m.bodyFields[m.cursor].Value)
				m.textInput.Focus()
				return m, textinput.Blink
			}
		}
	}

	return m, nil
}

func (m BodyEditorViewModel) View() string {
	if m.quitting {
		return ""
	}

	var view string

	view += "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += m.styles.Text("  Edit Request Body", m.styles.SelectedTitleColor) + "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += "\n"

	if len(m.bodyFields) == 0 {
		view += m.styles.Text("  No fields to edit", m.styles.ErrorColor) + "\n"
	} else {
		// Calculate visible range
		start := m.viewportStart
		end := m.viewportStart + m.maxVisible
		if end > len(m.bodyFields) {
			end = len(m.bodyFields)
		}

		// Render visible items
		for i := start; i < end; i++ {
			field := m.bodyFields[i]
			
			cursor := "  "
			if i == m.cursor {
				cursor = "► "
			}

			keyStyle := lipgloss.NewStyle().Foreground(m.styles.TitleColor).Bold(true)
			valueStyle := lipgloss.NewStyle().Foreground(m.styles.FooterColor)

			// If editing this field, show text input
			if m.editMode && i == m.editingIndex {
				view += cursor + keyStyle.Render(field.Key) + ": " + m.textInput.View() + "\n"
			} else {
				view += cursor + keyStyle.Render(field.Key) + ": " + valueStyle.Render(field.Value) + "\n"
			}
		}
	}

	view += "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += "\n"

	if m.editMode {
		view += m.styles.Text("ENTER to save • ESC to cancel", m.styles.FooterColor) + "\n"
	} else {
		view += m.styles.Text("↑↓ navigate • ENTER/E to edit • ESC to save & return • Q to cancel", m.styles.FooterColor) + "\n"
	}

	return view
}

func BodyEditorView(body interface{}, selected *string) {
	m := NewBodyEditorViewModel(body)
	m.selected = selected

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("BodyEditorView -> ", err)
		os.Exit(1)
	}
}
