package src

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (r *Runner) printRequestDetails(selectedRequest *RequestItem) {
	styles := DefaultStyles()

	fmt.Println()
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println(styles.Text(fmt.Sprintf("  Request: %s", selectedRequest.Name), styles.SelectedTitleColor))
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println()

	req := selectedRequest.Request

	// Method and URL
	methodColor := r.getMethodColor(req.Method, styles)
	fmt.Println(styles.Text(fmt.Sprintf("  Method:   %s", req.Method), methodColor))

	url := r.configLoader.ReplaceVariables(req.URL, r.config)
	fmt.Println(styles.Text(fmt.Sprintf("  URL:      %s", url), styles.FooterColor))
	fmt.Println()

	// Headers
	fmt.Println(styles.Text("  Headers:", styles.TitleColor))

	// Add JWT if not skipped
	if !req.SkipAuth {
		jwt := r.configLoader.GetJWT(r.secret)
		if jwt != "" {
			fmt.Println(styles.Text(fmt.Sprintf("    Authorization: Bearer %s", jwt), styles.AquamarineColor))
		}
	}

	// Global headers
	if r.config.GlobalHeaders != nil {
		for key, value := range r.config.GlobalHeaders {
			fmt.Println(styles.Text(fmt.Sprintf("    %s: %s", key, value), styles.MutedTitleColor))
		}
	}

	// Request-specific headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			fmt.Println(styles.Text(fmt.Sprintf("    %s: %s", key, value), styles.FooterColor))
		}
	}
	fmt.Println()

	// Body
	if req.Body != nil {
		fmt.Println(styles.Text("  Body:", styles.TitleColor))
		bodyJSON, _ := json.MarshalIndent(req.Body, "    ", "  ")
		fmt.Println(styles.Text(string(bodyJSON), styles.FooterColor))
	}

	fmt.Println()
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println()
}

func (r *Runner) printResponse(response *HTTPResponse, requestName string) {
	styles := DefaultStyles()

	fmt.Println()
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println(styles.Text(fmt.Sprintf("  Response: %s", requestName), styles.SelectedTitleColor))
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println()

	// Connection error
	if response.Error != nil {
		fmt.Println(styles.Text("  ❌ Error:", styles.ErrorColor))
		errorMsg := r.formatError(response.Error)
		fmt.Println(styles.Text("    "+errorMsg, styles.CoralColor))
		fmt.Println()
		fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
		fmt.Println()
		return
	}

	// Status Code
	statusColor := r.getStatusColor(response.StatusCode, styles)
	statusIcon := r.getStatusIcon(response.StatusCode)
	fmt.Println(styles.Text(fmt.Sprintf("  %s Status:   %d %s",
		statusIcon, response.StatusCode, response.Status), statusColor))

	// Duration
	durationColor := styles.FooterColor
	if response.Duration > 1*time.Second {
		durationColor = styles.CoralColor
	} else if response.Duration > 500*time.Millisecond {
		durationColor = styles.PeachColor
	} else {
		durationColor = styles.AquamarineColor
	}
	fmt.Println(styles.Text(fmt.Sprintf("  ⏱️  Duration: %s", response.Duration), durationColor))

	// Size
	fmt.Println(styles.Text(fmt.Sprintf("  📦 Size:     %s", formatBytes(response.Size)), styles.MutedTitleColor))
	fmt.Println()

	// Response headers
	fmt.Println(styles.Text("  📋 Response Headers:", styles.TitleColor))
	for key, values := range response.Headers {
		for _, value := range values {
			fmt.Println(styles.Text(fmt.Sprintf("    %s: %s", key, value), styles.FooterColor))
		}
	}
	fmt.Println()

	// Body
	if len(response.Body) == 0 {
		fmt.Println(styles.Text("  📄 Body: (empty)", styles.MutedTitleColor))
	} else {
		fmt.Println(styles.Text("  📄 Body:", styles.TitleColor))

		if response.IsJSON {
			prettyJSON, _ := json.MarshalIndent(response.BodyJSON, "    ", "  ")
			fmt.Println(styles.Text(string(prettyJSON), styles.FooterColor))
		} else {
			displayBody := response.BodyString
			if len(displayBody) > 1000 {
				displayBody = displayBody[:1000] + "... (truncated)"
			}
			fmt.Println(styles.Text("    "+displayBody, styles.FooterColor))
		}
	}

	fmt.Println()
	fmt.Println(styles.Text("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", styles.TitleColor))
	fmt.Println()
}

func (r *Runner) getStatusColor(statusCode int, styles *Styles) lipgloss.Color {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return styles.AquamarineColor
	case statusCode >= 300 && statusCode < 400:
		return styles.ThistleColor
	case statusCode >= 400 && statusCode < 500:
		return styles.CoralColor
	case statusCode >= 500:
		return styles.ErrorColor
	default:
		return styles.MutedTitleColor
	}
}

func (r *Runner) getStatusIcon(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "✓"
	case statusCode >= 300 && statusCode < 400:
		return "↪"
	case statusCode >= 400 && statusCode < 500:
		return "⚠"
	case statusCode >= 500:
		return "✗"
	default:
		return "?"
	}
}

func (r *Runner) formatError(err error) string {
	errStr := err.Error()

	if strings.Contains(errStr, "timeout") {
		return "⏱️  Request timeout - server took too long to respond"
	}

	if strings.Contains(errStr, "connection refused") {
		return "🚫 Connection refused - is the server running?"
	}

	if strings.Contains(errStr, "no such host") {
		return "🌐 DNS error - could not resolve hostname"
	}

	if strings.Contains(errStr, "network unreachable") {
		return "📡 Network unreachable - check your connection"
	}

	return errStr
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
