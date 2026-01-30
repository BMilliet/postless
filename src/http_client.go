package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	config       *ConfigJSON
	secret       *SecretJSON
	configLoader *ConfigLoader
}

type HTTPResponse struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       []byte
	BodyString string
	BodyJSON   interface{}
	IsJSON     bool
	Duration   time.Duration
	Size       int64
	Error      error
}

func NewHTTPClient(config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader) *HTTPClient {
	return &HTTPClient{
		config:       config,
		secret:       secret,
		configLoader: configLoader,
	}
}

func (c *HTTPClient) ExecuteRequest(request *RequestJSON) (*HTTPResponse, error) {
	response := &HTTPResponse{}

	// Start timing
	startTime := time.Now()

	// Interpolate URL
	url := c.configLoader.ReplaceVariables(request.URL, c.config)

	// Prepare body - ALWAYS create fresh buffer, never reuse
	var bodyReader io.Reader
	if request.Body != nil {
		// Marshal to JSON each time (no caching)
		bodyJSON, err := json.Marshal(request.Body)
		if err != nil {
			response.Error = fmt.Errorf("failed to marshal body: %v", err)
			return response, response.Error
		}

		// Create NEW buffer each time
		bodyReader = bytes.NewReader(bodyJSON)
	}

	// Create NEW request each time (no reuse)
	req, err := http.NewRequest(request.Method, url, bodyReader)
	if err != nil {
		response.Error = fmt.Errorf("failed to create request: %v", err)
		return response, response.Error
	}

	// ALWAYS set Content-Type for requests with body
	if request.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add headers
	c.addHeaders(req, request)

	// Get timeout from config (default: 30 seconds)
	timeout := time.Duration(c.config.GetTimeout()) * time.Second

	// Create NEW client each time with NO connection reuse
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)

	// Calculate duration
	response.Duration = time.Since(startTime)

	if err != nil {
		response.Error = err
		return response, err
	}
	defer resp.Body.Close()

	// Capture status
	response.StatusCode = resp.StatusCode
	response.Status = resp.Status

	// Capture headers
	response.Headers = resp.Header

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Errorf("failed to read body: %v", err)
		return response, response.Error
	}

	response.Body = body
	response.BodyString = string(body)
	response.Size = int64(len(body))

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		response.IsJSON = true
		response.BodyJSON = jsonData
	}

	return response, nil
}

func (c *HTTPClient) addHeaders(req *http.Request, request *RequestJSON) {
	// Global headers
	if c.config.GlobalHeaders != nil {
		for key, value := range c.config.GlobalHeaders {
			req.Header.Set(key, value)
		}
	}

	// JWT (if not skipAuth)
	jwt := c.configLoader.GetJWT(c.secret)
	if !request.SkipAuth && jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	// Request-specific headers (override previous)
	if request.Headers != nil {
		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}
	}
}
