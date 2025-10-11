package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading from various sources
type Loader struct {
	config      *Config
	configPath  string
	environment string
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		config:      &Config{},
		environment: getEnvOrDefault("APP_ENV", "development"),
	}
}

// Load loads configuration from all sources
func (l *Loader) Load(configPath string) (*Config, error) {
	l.configPath = configPath

	// 1. Set defaults
	if err := l.setDefaults(); err != nil {
		return nil, fmt.Errorf("setting defaults: %w", err)
	}

	// 2. Load from config file if exists
	if configPath != "" {
		if err := l.loadFromFile(configPath); err != nil {
			// Config file is optional
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("loading config file: %w", err)
			}
		}
	}

	// 3. Load environment-specific config
	if err := l.loadEnvironmentConfig(); err != nil {
		// Environment-specific config is optional
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading environment config: %w", err)
		}
	}

	// 4. Override with environment variables
	if err := l.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("loading from environment: %w", err)
	}

	// 5. Validate final configuration
	if err := l.config.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return l.config, nil
}

// LoadFromFile loads configuration from a specific file
func (l *Loader) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, l.config)
	case ".json":
		return json.Unmarshal(data, l.config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// loadEnvironmentConfig loads environment-specific configuration
func (l *Loader) loadEnvironmentConfig() error {
	if l.configPath == "" {
		return nil
	}

	dir := filepath.Dir(l.configPath)
	base := filepath.Base(l.configPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Try to load environment-specific config (e.g., config.production.yaml)
	envPath := filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, l.environment, ext))
	return l.loadFromFile(envPath)
}

// loadFromEnv loads configuration from environment variables
func (l *Loader) loadFromEnv() error {
	return l.walkStruct(reflect.ValueOf(l.config).Elem(), "")
}

// walkStruct recursively walks through struct fields and loads from env
func (l *Loader) walkStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Check for env tag
		envTag := field.Tag.Get("env")
		if envTag == "" && prefix == "" {
			// Recursively process nested structs
			if fieldValue.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Duration(0)) && field.Type != reflect.TypeOf(time.Time{}) {
				if err := l.walkStruct(fieldValue, field.Name); err != nil {
					return err
				}
			}
			continue
		}

		// Build environment variable name
		envName := envTag
		if envName == "" {
			envName = l.buildEnvName(prefix, field.Name)
		}

		// Get environment value
		envValue := os.Getenv(envName)
		if envValue == "" {
			continue
		}

		// Set the field value
		if err := l.setFieldValue(fieldValue, envValue); err != nil {
			return fmt.Errorf("setting field %s from env %s: %w", field.Name, envName, err)
		}
	}

	return nil
}

// buildEnvName builds environment variable name from struct path
func (l *Loader) buildEnvName(prefix, fieldName string) string {
	parts := []string{}
	if prefix != "" {
		parts = append(parts, l.toSnakeCase(prefix))
	}
	parts = append(parts, l.toSnakeCase(fieldName))
	return strings.ToUpper(strings.Join(parts, "_"))
}

// toSnakeCase converts CamelCase to snake_case
func (l *Loader) toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// setFieldValue sets a field value from string
func (l *Loader) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		// Handle string slices (comma-separated)
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			field.Set(reflect.ValueOf(values))
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// setDefaults sets default values from struct tags
func (l *Loader) setDefaults() error {
	return l.walkDefaults(reflect.ValueOf(l.config).Elem())
}

// walkDefaults recursively sets default values
func (l *Loader) walkDefaults(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Check for default tag
		defaultTag := field.Tag.Get("default")
		if defaultTag != "" && l.isZeroValue(fieldValue) {
			if err := l.setFieldValue(fieldValue, defaultTag); err != nil {
				return fmt.Errorf("setting default for field %s: %w", field.Name, err)
			}
		}

		// Recursively process nested structs
		if fieldValue.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Duration(0)) && field.Type != reflect.TypeOf(time.Time{}) {
			if err := l.walkDefaults(fieldValue); err != nil {
				return err
			}
		}
	}

	return nil
}

// isZeroValue checks if a value is zero
func (l *Loader) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	default:
		return false
	}
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustLoad loads configuration and panics on error
func MustLoad(configPath string) *Config {
	loader := NewLoader()
	config, err := loader.Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}

// LoadWithDefaults loads configuration with defaults only
func LoadWithDefaults() *Config {
	loader := NewLoader()
	config, _ := loader.Load("")
	return config
}
