package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadShellConfig(t *testing.T) {
	// Create a temp config file
	content := `# Patchman server
server=https://patchman.example.com

# API key for authentication
api_key=pm_test_key_12345

# Tags
tags=web,prod

# This should be ignored
protocol=2
unknown_key=ignored
`
	tmpFile, err := os.CreateTemp("", "patchman-client-*.conf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	tmpFile.Close()

	// Reset viper for this test
	viper.Reset()

	loadShellConfig(tmpFile.Name())

	if server := viper.GetString("server"); server != "https://patchman.example.com" {
		t.Errorf("Expected server 'https://patchman.example.com', got '%s'", server)
	}
	if apiKey := viper.GetString("api_key"); apiKey != "pm_test_key_12345" {
		t.Errorf("Expected api_key 'pm_test_key_12345', got '%s'", apiKey)
	}
	if tags := viper.GetString("tags"); tags != "web,prod" {
		t.Errorf("Expected tags 'web,prod', got '%s'", tags)
	}
}

func TestLoadShellConfigQuotedValues(t *testing.T) {
	content := `server="https://patchman.example.com"
api_key='pm_quoted_key'
tags="web, prod"
`
	tmpFile, err := os.CreateTemp("", "patchman-client-*.conf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	viper.Reset()
	loadShellConfig(tmpFile.Name())

	if server := viper.GetString("server"); server != "https://patchman.example.com" {
		t.Errorf("Expected unquoted server URL, got '%s'", server)
	}
	if apiKey := viper.GetString("api_key"); apiKey != "pm_quoted_key" {
		t.Errorf("Expected unquoted api_key, got '%s'", apiKey)
	}
}

func TestLoadShellConfigMissing(t *testing.T) {
	viper.Reset()
	// Should not panic on missing file
	loadShellConfig("/nonexistent/path/patchman-client.conf")
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pm_abcdef12345678", "pm_a****5678"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
	}

	for _, tt := range tests {
		result := MaskAPIKey(tt.input)
		if result != tt.expected {
			t.Errorf("MaskAPIKey(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}
