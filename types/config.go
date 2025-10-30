package types

// ConfigMap represents a parsed configuration file as key-value pairs
type ConfigMap map[string]string

// NewConfigMap creates a new ConfigMap
func NewConfigMap() ConfigMap {
	return make(ConfigMap)
}

// Set adds or updates a key-value pair in the configuration
func (cm ConfigMap) Set(key, value string) {
	cm[key] = value
}

// Get retrieves a value by key, returning empty string if not found
func (cm ConfigMap) Get(key string) string {
	return cm[key]
}

// Has checks if a key exists in the configuration
func (cm ConfigMap) Has(key string) bool {
	_, exists := cm[key]
	return exists
}
