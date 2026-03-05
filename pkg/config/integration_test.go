/*
   Copyright 2025 DurableCompose authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterComposeFiles(t *testing.T) {
	tests := []struct {
		name     string
		config   *DurableComposeConfig
		files    []string
		expected []string
	}{
		{
			name:     "nil config returns all files",
			config:   nil,
			files:    []string{"compose.yml", "compose.yaml"},
			expected: []string{"compose.yml", "compose.yaml"},
		},
		{
			name: "include specific files",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					IncludeFiles: []string{"docker-compose.yml", "compose.yml"},
				},
			},
			files:    []string{"docker-compose.yml", "compose.yml", "other.yml"},
			expected: []string{"docker-compose.yml", "compose.yml"},
		},
		{
			name: "exclude specific files",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					ExcludeFiles: []string{"*.override.yml"},
				},
			},
			files:    []string{"compose.yml", "compose.override.yml", "docker-compose.yml"},
			expected: []string{"compose.yml", "docker-compose.yml"},
		},
		{
			name: "file patterns",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					FilePatterns: []string{"*.compose.yml", "compose.*.yaml"},
				},
			},
			files:    []string{"web.compose.yml", "compose.dev.yaml", "other.yml"},
			expected: []string{"web.compose.yml", "compose.dev.yaml"},
		},
		{
			name: "enabled file types",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					EnabledFileTypes: []string{"yml"},
				},
			},
			files:    []string{"compose.yml", "compose.yaml"},
			expected: []string{"compose.yml"},
		},
		{
			name: "disabled file types",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					DisabledFileTypes: []string{"yaml"},
				},
			},
			files:    []string{"compose.yml", "compose.yaml"},
			expected: []string{"compose.yml"},
		},
		{
			name: "exclusions take precedence",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					IncludeFiles: []string{"*.yml"},
					ExcludeFiles: []string{"*.override.yml"},
				},
			},
			files:    []string{"compose.yml", "compose.override.yml"},
			expected: []string{"compose.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.FilterComposeFiles(tt.files)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldIncludeFile(t *testing.T) {
	config := &DurableComposeConfig{
		Loader: LoaderConfig{
			IncludeFiles:      []string{"docker-compose.yml"},
			ExcludeFiles:      []string{"*.override.yml"},
			EnabledFileTypes:  []string{"yml", "yaml"},
			DisabledFileTypes: []string{"json"},
		},
	}

	tests := []struct {
		file     string
		expected bool
	}{
		{"docker-compose.yml", true},
		{"compose.override.yml", false},
		{"compose.yml", false}, // Not in include list
		{"config.json", false},  // Disabled type
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			result := config.shouldIncludeFile(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSearchPaths(t *testing.T) {
	tests := []struct {
		name     string
		config   *DurableComposeConfig
		workDir  string
		expected []string
	}{
		{
			name:     "nil config returns workdir",
			config:   nil,
			workDir:  "/workspace",
			expected: []string{"/workspace"},
		},
		{
			name: "empty search paths returns workdir",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{},
			},
			workDir:  "/workspace",
			expected: []string{"/workspace"},
		},
		{
			name: "relative search paths",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					SearchPaths: []string{"./services", "./configs"},
				},
			},
			workDir:  "/workspace",
			expected: []string{"/workspace", "/workspace/services", "/workspace/configs"},
		},
		{
			name: "absolute search paths",
			config: &DurableComposeConfig{
				Loader: LoaderConfig{
					SearchPaths: []string{"/opt/services", "./local"},
				},
			},
			workDir:  "/workspace",
			expected: []string{"/workspace", "/opt/services", "/workspace/local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetSearchPaths(tt.workDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetServiceEnvironment(t *testing.T) {
	config := &DurableComposeConfig{
		Services: map[string]ServiceConfig{
			"web": {
				Environment: map[string]string{
					"PORT":  "8080",
					"DEBUG": "true",
				},
			},
		},
	}

	tests := []struct {
		name        string
		serviceName string
		expected    map[string]string
	}{
		{
			name:        "existing service",
			serviceName: "web",
			expected: map[string]string{
				"PORT":  "8080",
				"DEBUG": "true",
			},
		},
		{
			name:        "non-existing service",
			serviceName: "api",
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetServiceEnvironment(tt.serviceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetServiceBuildConfig(t *testing.T) {
	config := &DurableComposeConfig{
		Services: map[string]ServiceConfig{
			"web": {
				Build: &BuildConfig{
					Context:    "./web",
					Dockerfile: "Dockerfile.dev",
					Args: map[string]string{
						"NODE_VERSION": "18",
					},
				},
			},
			"api": {},
		},
	}

	tests := []struct {
		name        string
		serviceName string
		expected    *BuildConfig
	}{
		{
			name:        "service with build config",
			serviceName: "web",
			expected: &BuildConfig{
				Context:    "./web",
				Dockerfile: "Dockerfile.dev",
				Args: map[string]string{
					"NODE_VERSION": "18",
				},
			},
		},
		{
			name:        "service without build config",
			serviceName: "api",
			expected:    nil,
		},
		{
			name:        "non-existing service",
			serviceName: "db",
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetServiceBuildConfig(tt.serviceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
