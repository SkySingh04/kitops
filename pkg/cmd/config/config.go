// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"kitops/pkg/output"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Config struct {
	LogLevel  string `json:"log_level"`
	Progress  string `json:"progress"`
	ConfigDir string `json:"config_dir"`
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  output.LogLevelInfo.String(),
		Progress:  "plain",
		ConfigDir: "",
	}
}

// Set a configuration key and value.
func setConfig(_ context.Context, opts *configOptions) error {
	configPath := getConfigPath(opts.profile)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		cfg = DefaultConfig() // Start with defaults if config doesn't exist.
	}

	v := reflect.ValueOf(cfg).Elem().FieldByName(strings.Title(opts.key))
	if !v.IsValid() {
		return fmt.Errorf("unknown configuration key: %s", opts.key)
	}

	v.SetString(opts.value)
	err = SaveConfig(cfg, configPath)
	if err != nil {
		return err
	}
	fmt.Printf("Config '%s' set to '%s'\n", opts.key, opts.value)
	return nil
}

// Get a configuration value.
func getConfig(_ context.Context, opts *configOptions) (string, error) {
	configPath := getConfigPath(opts.profile)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return "", err
	}

	v := reflect.ValueOf(cfg).Elem().FieldByName(strings.Title(opts.key))
	if !v.IsValid() {
		return "", fmt.Errorf("unknown configuration key: %s", opts.key)
	}

	return fmt.Sprintf("%v", v.Interface()), nil
}

// List all configuration values.
func listConfig(_ context.Context, opts *configOptions) error {
	configPath := getConfigPath(opts.profile)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	// Use reflection to iterate through fields and print them.
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s: %v\n", t.Field(i).Name, v.Field(i).Interface())
	}
	return nil
}

// Reset configuration to defaults.
func resetConfig(_ context.Context, opts *configOptions) error {
	configPath := getConfigPath(opts.profile)
	cfg := DefaultConfig()
	err := SaveConfig(cfg, configPath)
	if err != nil {
		return err
	}
	fmt.Println("Configuration reset to default values.")
	return nil
}

// Load configuration from a file.
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil // Return default config if file doesn't exist.
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Save configuration to a file.
func SaveConfig(config *Config, configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(config)
}

// Get the config path, either from the profile or default.
func getConfigPath(profile string) string {
	configDir := os.Getenv("KITOPS_HOME")
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".kitops")
	}
	if profile != "" {
		configDir = filepath.Join(configDir, "profiles", profile)
	}
	return filepath.Join(configDir, "config.json")
}

// ConfigOptions struct to store command options.
type configOptions struct {
	key        string
	value      string
	profile    string
	configHome string
}
