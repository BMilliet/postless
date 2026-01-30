package src

import (
	"encoding/json"
	"fmt"
)

func ParseJSONContent[T any](content string) (*T, error) {
	var result T
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("ParseJSONContent -> %v", err)
	}
	return &result, nil
}

func ToJSON(v interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("ToJSON -> %v", err)
	}
	return string(jsonBytes), nil
}
