package src

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RequestPreviewViewModel struct {
	selectedRequest *RequestItem
	config          *ConfigJSON
	secret          *SecretJSON
	configLoader    *ConfigLoader
	action          *string
	quitting        bool
	styles          *Styles
}

func NewRequestPreviewViewModel(selectedRequest *RequestItem, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader) RequestPreviewViewModel {
	return RequestPreviewViewModel{
		selectedRequest: selectedRequest,
		config:          config,
		secret:          secret,
		configLoader:    configLoader,
		quitting:        false,
		styles:          DefaultStyles(),
	}
}

func (m RequestPreviewViewModel) Init() tea.Cmd {
	return nil
}

func (m RequestPreviewViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			*m.action = "cancel"
			m.quitting = true
			return m, tea.Quit
		case "e", "E":
			*m.action = "edit"
			m.quitting = true
			return m, tea.Quit
		case "enter":
			*m.action = "execute"
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m RequestPreviewViewModel) View() string {
	if m.quitting {
		return ""
	}

	req := m.selectedRequest.Request

	// Build the view
	var view string

	view += "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += m.styles.Text(fmt.Sprintf("  Request: %s", m.selectedRequest.Name), m.styles.SelectedTitleColor) + "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += "\n"

	// Method and URL
	methodColor := getMethodColor(req.Method, m.styles)
	view += m.styles.Text(fmt.Sprintf("  Method:   %s", req.Method), methodColor) + "\n"

	url := m.configLoader.ReplaceVariables(req.URL, m.config)
	view += m.styles.Text(fmt.Sprintf("  URL:      %s", url), m.styles.FooterColor) + "\n"
	view += "\n"

	// Headers
	view += m.styles.Text("  Headers:", m.styles.TitleColor) + "\n"

	// Add JWT if not skipped
	if !req.SkipAuth {
		jwt := m.configLoader.GetJWT(m.secret)
		if jwt != "" {
			view += m.styles.Text(fmt.Sprintf("    Authorization: Bearer %s", jwt), m.styles.AquamarineColor) + "\n"
		}
	}

	// Global headers
	if m.config.GlobalHeaders != nil {
		for key, value := range m.config.GlobalHeaders {
			view += m.styles.Text(fmt.Sprintf("    %s: %s", key, value), m.styles.MutedTitleColor) + "\n"
		}
	}

	// Request-specific headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			view += m.styles.Text(fmt.Sprintf("    %s: %s", key, value), m.styles.FooterColor) + "\n"
		}
	}
	view += "\n"

	// Body
	if req.Body != nil {
		view += m.styles.Text("  Body:", m.styles.TitleColor) + "\n"
		bodyJSON, _ := json.MarshalIndent(req.Body, "    ", "  ")
		view += m.styles.Text(string(bodyJSON), m.styles.FooterColor) + "\n"
	}

	view += "\n"
	view += m.styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", m.styles.TitleColor) + "\n"
	view += "\n"

	// Footer with instructions
	view += m.styles.Text("Press ENTER to execute • E to edit body • Q/ESC to cancel", m.styles.FooterColor) + "\n"

	return view
}

func getMethodColor(method string, styles *Styles) lipgloss.Color {
	switch method {
	case "GET":
		return styles.AquamarineColor
	case "POST":
		return styles.PeachColor
	case "PUT":
		return styles.ThistleColor
	case "DELETE":
		return styles.CoralColor
	case "PATCH":
		return styles.OrchidColor
	default:
		return styles.FooterColor
	}
}

func RequestPreviewView(selectedRequest *RequestItem, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader, action *string) {
	m := NewRequestPreviewViewModel(selectedRequest, config, secret, configLoader)
	m.action = action

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("RequestPreviewView -> ", err)
		os.Exit(1)
	}
}
