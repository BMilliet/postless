package src

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Runner struct {
	fileManager  FileManagerInterface
	utils        UtilsInterface
	viewBuilder  ViewBuilderInterface
	configLoader *ConfigLoader
	config       *ConfigJSON
	secret       *SecretJSON
}

func NewRunner(fm FileManagerInterface, u UtilsInterface, b ViewBuilderInterface) *Runner {
	return &Runner{
		fileManager:  fm,
		utils:        u,
		viewBuilder:  b,
		configLoader: NewConfigLoader(fm),
	}
}

func (r *Runner) Start() {
	styles := DefaultStyles()

	// Step 1: Check if postless directory exists
	exists, err := r.fileManager.CheckPostlessDir()
	if err != nil {
		r.utils.HandleError(err, "Failed to check postless directory")
	}

	if !exists {
		fmt.Println(styles.Text("⚠️  Postless directory not found in current location", styles.ErrorColor))
		return
	}

	// Step 2: Check if config.json exists
	configExists, err := r.fileManager.CheckConfigYML()
	if err != nil {
		r.utils.HandleError(err, "Failed to check config.json")
	}

	if !configExists {
		fmt.Println(styles.Text("⚠️  config.json not found in postless directory", styles.ErrorColor))
		return
	}

	// Step 3: Load and validate config.json
	config, err := r.configLoader.LoadConfigJSON()
	if err != nil {
		fmt.Println(styles.Text("⚠️  Invalid config.json: "+err.Error(), styles.ErrorColor))
		return
	}
	r.config = config

	// Step 4: Load or create secret.json
	secret, err := r.configLoader.LoadSecretJSON()
	if err != nil {
		fmt.Println(styles.Text("⚠️  Failed to load secret.json: "+err.Error(), styles.ErrorColor))
		return
	}
	r.secret = secret

	// Step 5: Check if requests directory exists
	requestsExists, err := r.fileManager.CheckRequestsDir()
	if err != nil {
		r.utils.HandleError(err, "Failed to check requests directory")
	}

	if !requestsExists {
		fmt.Println(styles.Text("⚠️  No requests directory found", styles.ErrorColor))
		return
	}

	// Step 6: Load collections
	collections, err := r.configLoader.LoadCollections()
	if err != nil {
		r.utils.HandleError(err, "Failed to load collections")
	}

	if len(collections) == 0 {
		fmt.Println(styles.Text("⚠️  No collections found in requests directory", styles.ErrorColor))
		return
	}

	// Step 7: Show collections view
	result := r.viewBuilder.NewCollectionsView(collections, r.config, r.secret, r.configLoader, r.fileManager)
	r.utils.ValidateInput(result)

	// Parse result: "collection|requestName" or "settings|key"
	parts := strings.Split(result, "|")
	if len(parts) != 2 {
		return
	}

	pageType := parts[0]
	itemName := parts[1]

	// Handle settings
	if pageType == "settings" {
		r.handleSettings(itemName)
		return
	}

	// Handle regular request
	collectionName := pageType
	requestName := itemName

	// Find the selected request
	var selectedRequest *RequestItem
	for i := range collections {
		if collections[i].Name == collectionName {
			for j := range collections[i].Requests {
				if collections[i].Requests[j].Name == requestName {
					selectedRequest = &collections[i].Requests[j]
					break
				}
			}
			break
		}
	}

	if selectedRequest == nil {
		return
	}

	// Loop to allow body editing
	for {
		// Show request preview and get user action
		clearScreen()
		action := r.viewBuilder.NewRequestPreviewView(selectedRequest, r.config, r.secret, r.configLoader)

		if action == "cancel" {
			return
		}

		if action == "edit" {
			// Edit body
			if selectedRequest.Request.Body == nil {
				clearScreen()
				fmt.Println()
				fmt.Println(styles.Text("⚠️  This request has no body to edit", styles.ErrorColor))
				fmt.Println()
				fmt.Println(styles.Text("Press ENTER to continue...", styles.FooterColor))
				fmt.Scanln()
				continue
			}

			clearScreen()
			result := r.viewBuilder.NewBodyEditorView(selectedRequest.Request.Body)

			if result == ExitSignal {
				continue // Back to preview (user pressed Q)
			}

			// Parse result: "key1|value1||key2|value2||..."
			if result == "" {
				continue
			}

			// Update all fields
			fieldPairs := strings.Split(result, "||")
			if bodyMap, ok := selectedRequest.Request.Body.(map[string]interface{}); ok {
				for _, pair := range fieldPairs {
					parts := strings.Split(pair, "|")
					if len(parts) == 2 {
						fieldKey := parts[0]
						newValue := parts[1]

						// Try to parse as number
						var parsedValue interface{}
						var numValue float64
						if _, err := fmt.Sscanf(newValue, "%f", &numValue); err == nil {
							parsedValue = numValue
						} else if newValue == "true" || newValue == "false" {
							parsedValue = (newValue == "true")
						} else {
							parsedValue = newValue
						}

						bodyMap[fieldKey] = parsedValue
					}
				}

				// Create a NEW map to ensure no reference issues
				newBodyMap := make(map[string]interface{})
				for k, v := range bodyMap {
					newBodyMap[k] = v
				}
				selectedRequest.Request.Body = newBodyMap

				// Save changes to file
				err := r.fileManager.SaveRequestJSON(selectedRequest.FilePath, selectedRequest.Request)
				if err != nil {
					clearScreen()
					fmt.Println()
					fmt.Println(styles.Text(fmt.Sprintf("⚠️  Failed to save changes: %v", err), styles.ErrorColor))
					fmt.Println()
					fmt.Println(styles.Text("Press ENTER to continue...", styles.FooterColor))
					fmt.Scanln()
					continue
				}

				// Reload the request from file to ensure consistency
				content, err := r.fileManager.ReadFileContent(selectedRequest.FilePath)
				if err == nil {
					reloadedRequest, err := ParseJSONContent[RequestJSON](content)
					if err == nil {
						selectedRequest.Request = reloadedRequest
					}
				}
			}

			continue // Back to preview with updated values
		}

		// If action is "execute", break the loop
		if action == "execute" {
			break
		}
	}

	// Clear screen after confirmation
	clearScreen()

	// Execute request
	fmt.Println()
	fmt.Println(styles.Text("⏳ Executing request...", styles.ThistleColor))
	fmt.Println()

	httpClient := NewHTTPClient(r.config, r.secret, r.configLoader)
	response, _ := httpClient.ExecuteRequest(selectedRequest.Request)

	// Clear the "executing" message and print response
	clearScreen()
	r.printResponse(response, selectedRequest.Name)
}

func (r *Runner) getMethodColor(method string, styles *Styles) lipgloss.Color {
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
		return styles.MutedTitleColor
	}
}

// clearScreen clears the terminal screen
func clearScreen() {
	// ANSI escape code to clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
}

func (r *Runner) handleSettings(settingKey string) {
	styles := DefaultStyles()

	var currentValue string
	var prompt string

	switch settingKey {
	case "baseUrl":
		currentValue = r.config.BaseUrl
		prompt = fmt.Sprintf("Current Base URL: %s\nEnter new Base URL (or press ESC to cancel):", currentValue)
	case "jwt":
		currentValue = r.configLoader.GetJWT(r.secret)
		prompt = fmt.Sprintf("Current JWT: %s\nEnter new JWT token (or press ESC to cancel):", currentValue)
	case "timeout":
		currentValue = fmt.Sprintf("%d", r.config.GetTimeout())
		prompt = fmt.Sprintf("Current Timeout: %s seconds\nEnter new timeout in seconds (or press ESC to cancel):", currentValue)
	default:
		return
	}

	// Get new value from user
	newValue := r.viewBuilder.NewTextFieldView(prompt, currentValue)

	if newValue == ExitSignal || newValue == "" {
		return
	}

	// Update config or secret
	switch settingKey {
	case "baseUrl":
		r.config.BaseUrl = newValue
	case "jwt":
		r.secret.JWT = strings.TrimSpace(newValue)
		// Save secret.json
		secretJSON, err := ToJSON(r.secret)
		if err != nil {
			fmt.Println(styles.Text("Failed to serialize secret: "+err.Error(), styles.ErrorColor))
			return
		}
		if err := r.fileManager.WriteSecretContent(secretJSON); err != nil {
			fmt.Println(styles.Text("Failed to save secret: "+err.Error(), styles.ErrorColor))
			return
		}
	case "timeout":
		// Parse timeout
		var timeout int
		fmt.Sscanf(newValue, "%d", &timeout)
		if timeout > 0 {
			r.config.Timeout = timeout
		}
	}

	// Save config if baseUrl or timeout changed
	if settingKey == "baseUrl" || settingKey == "timeout" {
		configJSON, err := ToJSON(r.config)
		if err != nil {
			fmt.Println(styles.Text("Failed to serialize config: "+err.Error(), styles.ErrorColor))
			return
		}

		if err := r.fileManager.WriteConfigContent(configJSON); err != nil {
			fmt.Println(styles.Text("Failed to save config: "+err.Error(), styles.ErrorColor))
			return
		}
	}

	fmt.Println()
	fmt.Println(styles.Text("✓ Settings updated successfully!", styles.AquamarineColor))
	fmt.Println()
}
