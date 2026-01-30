package src

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type ConfigLoader struct {
	fileManager FileManagerInterface
}

func NewConfigLoader(fm FileManagerInterface) *ConfigLoader {
	return &ConfigLoader{fileManager: fm}
}

func (cl *ConfigLoader) LoadConfigJSON() (*ConfigJSON, error) {
	content, err := cl.fileManager.GetConfigContent()
	if err != nil {
		return nil, fmt.Errorf("LoadConfigJSON -> %v", err)
	}

	var config ConfigJSON
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("LoadConfigJSON -> failed to parse JSON: %v", err)
	}

	// Validate required fields
	if config.BaseUrl == "" {
		return nil, fmt.Errorf("LoadConfigJSON -> baseUrl is required")
	}

	return &config, nil
}

func (cl *ConfigLoader) LoadCollections() ([]Collection, error) {
	fm := cl.fileManager.(*FileManager)

	collectionNames, err := cl.fileManager.GetCollections()
	if err != nil {
		return nil, fmt.Errorf("LoadCollections -> %v", err)
	}

	var collections []Collection
	for _, collName := range collectionNames {
		collection := Collection{
			Name:     collName,
			Path:     filepath.Join(fm.RequestsDir, collName),
			Requests: []RequestItem{},
		}

		files, err := cl.fileManager.GetRequestFiles(collName)
		if err != nil {
			return nil, fmt.Errorf("LoadCollections -> failed to get files for %s: %v", collName, err)
		}

		for _, file := range files {
			filePath := filepath.Join(collection.Path, file)
			content, err := cl.fileManager.ReadFileContent(filePath)
			if err != nil {
				continue // Skip files that can't be read
			}

			var request RequestJSON
			if err := json.Unmarshal([]byte(content), &request); err != nil {
				continue // Skip files that can't be parsed
			}

			requestItem := RequestItem{
				Name:     request.Name,
				FileName: file,
				FilePath: filePath,
				Request:  &request,
			}

			collection.Requests = append(collection.Requests, requestItem)
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

func (cl *ConfigLoader) LoadSecretJSON() (*SecretJSON, error) {
	// Try to load existing secret.json
	content, err := cl.fileManager.GetSecretContent()
	if err != nil {
		// If file doesn't exist, create default
		defaultSecret := GetDefaultSecretJSON()
		jsonContent, err := ToJSON(defaultSecret)
		if err != nil {
			return nil, fmt.Errorf("LoadSecretJSON -> failed to create default: %v", err)
		}

		err = cl.fileManager.WriteSecretContent(jsonContent)
		if err != nil {
			return nil, fmt.Errorf("LoadSecretJSON -> failed to write default: %v", err)
		}

		return defaultSecret, nil
	}

	var secret SecretJSON
	if err := json.Unmarshal([]byte(content), &secret); err != nil {
		return nil, fmt.Errorf("LoadSecretJSON -> failed to parse JSON: %v", err)
	}

	return &secret, nil
}

func (cl *ConfigLoader) GetBaseURL(config *ConfigJSON) string {
	return config.BaseUrl
}

func (cl *ConfigLoader) GetJWT(secret *SecretJSON) string {
	if secret == nil {
		return ""
	}
	return strings.TrimSpace(secret.JWT)
}

func (cl *ConfigLoader) ReplaceVariables(text string, config *ConfigJSON) string {
	result := text

	// Replace {{baseUrl}}
	baseUrl := cl.GetBaseURL(config)
	result = strings.ReplaceAll(result, "{{baseUrl}}", baseUrl)

	return result
}
