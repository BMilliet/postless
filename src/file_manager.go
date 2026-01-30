package src

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileManagerInterface interface {
	CheckIfPathExists(path string) (bool, error)
	ReadFileContent(filePath string) (string, error)
	WriteFileContent(filePath, content string) error
	GetConfigContent() (string, error)
	WriteConfigContent(content string) error
	GetSecretContent() (string, error)
	WriteSecretContent(content string) error
	CheckPostlessDir() (bool, error)
	CheckConfigYML() (bool, error)
	CheckSecretJSON() (bool, error)
	CheckRequestsDir() (bool, error)
	GetCollections() ([]string, error)
	GetRequestFiles(collectionName string) ([]string, error)
	GetCurrentDirectoryName() (string, error)
	SaveRequestJSON(filePath string, request *RequestJSON) error
}

type FileManager struct {
	CurrentDir        string
	PostlessDir       string
	ConfigPath        string
	SecretPath        string
	RequestsDir       string
	PostlessDirExists bool
}

func NewFileManager() (*FileManager, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("NewFileManager -> failed to get current directory: %v", err)
	}

	postlessDir := filepath.Join(currentDir, PostlessDirName)
	configPath := filepath.Join(postlessDir, ConfigFileName)
	secretPath := filepath.Join(postlessDir, SecretFileName)
	requestsDir := filepath.Join(postlessDir, RequestsDirName)

	return &FileManager{
		CurrentDir:        currentDir,
		PostlessDir:       postlessDir,
		ConfigPath:        configPath,
		SecretPath:        secretPath,
		RequestsDir:       requestsDir,
		PostlessDirExists: false,
	}, nil
}

func (m *FileManager) CheckPostlessDir() (bool, error) {
	exists, err := m.CheckIfPathExists(m.PostlessDir)
	if err != nil {
		return false, fmt.Errorf("CheckPostlessDir -> %v", err)
	}

	if exists {
		info, err := os.Stat(m.PostlessDir)
		if err != nil {
			return false, fmt.Errorf("CheckPostlessDir -> failed to stat: %v", err)
		}
		if !info.IsDir() {
			return false, fmt.Errorf("CheckPostlessDir -> '%s' exists but is not a directory", PostlessDirName)
		}
	}

	m.PostlessDirExists = exists
	return exists, nil
}

func (m *FileManager) CheckConfigYML() (bool, error) {
	exists, err := m.CheckIfPathExists(m.ConfigPath)
	if err != nil {
		return false, fmt.Errorf("CheckConfigYML -> %v", err)
	}
	return exists, nil
}

func (m *FileManager) CheckSecretJSON() (bool, error) {
	exists, err := m.CheckIfPathExists(m.SecretPath)
	if err != nil {
		return false, fmt.Errorf("CheckSecretJSON -> %v", err)
	}
	return exists, nil
}

func (m *FileManager) CheckRequestsDir() (bool, error) {
	exists, err := m.CheckIfPathExists(m.RequestsDir)
	if err != nil {
		return false, fmt.Errorf("CheckRequestsDir -> %v", err)
	}

	if exists {
		info, err := os.Stat(m.RequestsDir)
		if err != nil {
			return false, fmt.Errorf("CheckRequestsDir -> failed to stat: %v", err)
		}
		if !info.IsDir() {
			return false, fmt.Errorf("CheckRequestsDir -> '%s' exists but is not a directory", RequestsDirName)
		}
	}

	return exists, nil
}

func (m *FileManager) GetCollections() ([]string, error) {
	entries, err := os.ReadDir(m.RequestsDir)
	if err != nil {
		return nil, fmt.Errorf("GetCollections -> failed to read directory: %v", err)
	}

	var collections []string
	for _, entry := range entries {
		if entry.IsDir() {
			collections = append(collections, entry.Name())
		}
	}

	return collections, nil
}

func (m *FileManager) GetRequestFiles(collectionName string) ([]string, error) {
	collectionPath := filepath.Join(m.RequestsDir, collectionName)
	entries, err := os.ReadDir(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("GetRequestFiles -> failed to read directory: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (m *FileManager) CheckIfPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("CheckIfPathExists -> %v", err)
}

func (m *FileManager) ReadFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("ReadFileContent -> %s %v", filePath, err)
	}
	return string(data), nil
}

func (m *FileManager) WriteFileContent(filePath, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("WriteFileContent -> %s %v", filePath, err)
	}
	return nil
}

func (m *FileManager) GetConfigContent() (string, error) {
	str, err := m.ReadFileContent(m.ConfigPath)
	if err != nil {
		return "", fmt.Errorf("GetConfigContent -> %s %v", m.ConfigPath, err)
	}
	return str, nil
}

func (m *FileManager) WriteConfigContent(content string) error {
	err := m.WriteFileContent(m.ConfigPath, content)
	if err != nil {
		return fmt.Errorf("WriteConfigContent -> %s: %v", m.ConfigPath, err)
	}
	return nil
}

func (m *FileManager) GetSecretContent() (string, error) {
	str, err := m.ReadFileContent(m.SecretPath)
	if err != nil {
		return "", fmt.Errorf("GetSecretContent -> %s %v", m.SecretPath, err)
	}
	return str, nil
}

func (m *FileManager) WriteSecretContent(content string) error {
	err := m.WriteFileContent(m.SecretPath, content)
	if err != nil {
		return fmt.Errorf("WriteSecretContent -> %s: %v", m.SecretPath, err)
	}
	return nil
}

func (m *FileManager) GetCurrentDirectoryName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("GetCurrentDirectoryName -> %v", err)
	}

	return filepath.Base(dir), nil
}

func (m *FileManager) SaveRequestJSON(filePath string, request *RequestJSON) error {
	content, err := ToJSON(request)
	if err != nil {
		return fmt.Errorf("SaveRequestJSON -> failed to marshal: %v", err)
	}

	err = m.WriteFileContent(filePath, content)
	if err != nil {
		return fmt.Errorf("SaveRequestJSON -> %v", err)
	}

	return nil
}
