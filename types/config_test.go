package types

import (
	"testing"
)

func TestNewConfigMap(t *testing.T) {
	config := NewConfigMap()
	
	if config == nil {
		t.Fatal("NewConfigMap() returned nil")
	}
	
	if len(config) != 0 {
		t.Errorf("NewConfigMap() length = %v, want 0", len(config))
	}
}

func TestConfigMap_Set(t *testing.T) {
	config := NewConfigMap()
	
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "set title",
			key:   "title",
			value: "My Game",
		},
		{
			name:  "set width",
			key:   "width",
			value: "1280",
		},
		{
			name:  "set height",
			key:   "height",
			value: "720",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Set(tt.key, tt.value)
			
			if config[tt.key] != tt.value {
				t.Errorf("ConfigMap.Set() failed, config[%v] = %v, want %v", tt.key, config[tt.key], tt.value)
			}
		})
	}
}

func TestConfigMap_Get(t *testing.T) {
	config := NewConfigMap()
	config["title"] = "My Game"
	config["width"] = "1280"
	
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "get existing key",
			key:      "title",
			expected: "My Game",
		},
		{
			name:     "get another existing key",
			key:      "width",
			expected: "1280",
		},
		{
			name:     "get non-existing key",
			key:      "height",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.Get(tt.key)
			if result != tt.expected {
				t.Errorf("ConfigMap.Get(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestConfigMap_Has(t *testing.T) {
	config := NewConfigMap()
	config["title"] = "My Game"
	config["width"] = "1280"
	
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "has existing key",
			key:      "title",
			expected: true,
		},
		{
			name:     "has another existing key",
			key:      "width",
			expected: true,
		},
		{
			name:     "has non-existing key",
			key:      "height",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.Has(tt.key)
			if result != tt.expected {
				t.Errorf("ConfigMap.Has(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}
