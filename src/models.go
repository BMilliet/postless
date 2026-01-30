package src

type ConfigJSON struct {
	BaseUrl       string            `json:"baseUrl"`
	Timeout       int               `json:"timeout,omitempty"` // Timeout in seconds (optional, default: 30)
	GlobalHeaders map[string]string `json:"globalHeaders,omitempty"`
}

type SecretJSON struct {
	JWT string `json:"jwt"`
}

type RequestJSON struct {
	Name     string            `json:"name"`
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	SkipAuth bool              `json:"skipAuth"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
}

type Collection struct {
	Name     string
	Path     string
	Requests []RequestItem
}

type RequestItem struct {
	Name     string
	FileName string
	FilePath string
	Request  *RequestJSON
}

func GetDefaultConfigJSON() *ConfigJSON {
	return &ConfigJSON{
		BaseUrl: "http://localhost:3000",
		Timeout: 30,
		GlobalHeaders: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

func GetDefaultSecretJSON() *SecretJSON {
	return &SecretJSON{
		JWT: "",
	}
}

// GetTimeout returns the configured timeout or default (30 seconds)
func (c *ConfigJSON) GetTimeout() int {
	if c.Timeout <= 0 {
		return 30 // Default timeout: 30 seconds
	}
	return c.Timeout
}
