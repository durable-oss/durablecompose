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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDurableComposeConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "durablecompose.toml")

	configContent := `
[project]
name = "test-project"
profiles = ["dev", "test"]
work_dir = "/workspace"
env_files = [".env", ".env.local"]

[services.web]
extends = "base-web"
extends_file = "base.toml"

[services.web.environment]
PORT = "8080"
DEBUG = "true"

[services.web.build]
context = "./web"
dockerfile = "Dockerfile.dev"
target = "development"

[services.web.build.args]
NODE_VERSION = "18"

[loader]
enabled_file_types = ["yaml", "yml"]
disabled_file_types = ["json"]
file_patterns = ["*.compose.yml", "compose.*.yaml"]
include_files = ["docker-compose.yml"]
exclude_files = ["docker-compose.override.yml"]
search_paths = ["./services", "./configs"]

[builders.custom]
type = "buildkit"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test loading
	config, err := LoadDurableComposeConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify project settings
	assert.Equal(t, "test-project", config.Project.Name)
	assert.Equal(t, []string{"dev", "test"}, config.Project.Profiles)
	assert.Equal(t, "/workspace", config.Project.WorkDir)
	assert.Equal(t, []string{".env", ".env.local"}, config.Project.EnvFiles)

	// Verify service config
	require.Contains(t, config.Services, "web")
	webSvc := config.Services["web"]
	assert.Equal(t, "base-web", webSvc.Extends)
	assert.Equal(t, "base.toml", webSvc.ExtendsFile)
	assert.Equal(t, "8080", webSvc.Environment["PORT"])
	assert.Equal(t, "true", webSvc.Environment["DEBUG"])

	// Verify build config
	require.NotNil(t, webSvc.Build)
	assert.Equal(t, "./web", webSvc.Build.Context)
	assert.Equal(t, "Dockerfile.dev", webSvc.Build.Dockerfile)
	assert.Equal(t, "development", webSvc.Build.Target)
	assert.Equal(t, "18", webSvc.Build.Args["NODE_VERSION"])

	// Verify loader config
	assert.Equal(t, []string{"yaml", "yml"}, config.Loader.EnabledFileTypes)
	assert.Equal(t, []string{"json"}, config.Loader.DisabledFileTypes)
	assert.Equal(t, []string{"*.compose.yml", "compose.*.yaml"}, config.Loader.FilePatterns)
	assert.Equal(t, []string{"docker-compose.yml"}, config.Loader.IncludeFiles)
	assert.Equal(t, []string{"docker-compose.override.yml"}, config.Loader.ExcludeFiles)
	assert.Equal(t, []string{"./services", "./configs"}, config.Loader.SearchPaths)

	// Verify builders
	require.Contains(t, config.Builders, "custom")
	assert.Equal(t, "buildkit", config.Builders["custom"].Type)
}

func TestLoadDurableComposeConfigFromDir(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create config in parent directory
	configPath := filepath.Join(tmpDir, "durablecompose.toml")
	configContent := `
[project]
name = "found-project"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test loading from subdirectory
	config, err := LoadDurableComposeConfigFromDir(subDir)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "found-project", config.Project.Name)
}

func TestFindDurableComposeConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindDurableComposeConfig(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDurableComposeConfig_Merge(t *testing.T) {
	base := &DurableComposeConfig{
		Project: ProjectConfig{
			Name:     "base-project",
			Profiles: []string{"base"},
		},
		Services: map[string]ServiceConfig{
			"web": {
				Environment: map[string]string{
					"BASE_VAR": "base",
				},
			},
		},
		Loader: LoaderConfig{
			EnabledFileTypes: []string{"yaml"},
		},
	}

	override := &DurableComposeConfig{
		Project: ProjectConfig{
			Name:     "override-project",
			Profiles: []string{"override"},
			WorkDir:  "/override",
		},
		Services: map[string]ServiceConfig{
			"web": {
				Environment: map[string]string{
					"OVERRIDE_VAR": "override",
				},
			},
			"api": {
				Environment: map[string]string{
					"API_VAR": "api",
				},
			},
		},
		Loader: LoaderConfig{
			EnabledFileTypes: []string{"yml"},
		},
	}

	base.Merge(override)

	// Verify merged project settings
	assert.Equal(t, "override-project", base.Project.Name)
	assert.Equal(t, []string{"base", "override"}, base.Project.Profiles)
	assert.Equal(t, "/override", base.Project.WorkDir)

	// Verify merged services
	require.Contains(t, base.Services, "web")
	require.Contains(t, base.Services, "api")
	assert.Equal(t, "base", base.Services["web"].Environment["BASE_VAR"])
	assert.Equal(t, "override", base.Services["web"].Environment["OVERRIDE_VAR"])
	assert.Equal(t, "api", base.Services["api"].Environment["API_VAR"])

	// Verify merged loader config
	assert.Equal(t, []string{"yml"}, base.Loader.EnabledFileTypes)
}

func TestServiceConfig_MergeFrom(t *testing.T) {
	base := ServiceConfig{
		Extends: "base",
		Environment: map[string]string{
			"VAR1": "value1",
		},
		Build: &BuildConfig{
			Context: "./base",
			Args: map[string]string{
				"ARG1": "arg1",
			},
		},
	}

	override := ServiceConfig{
		Extends: "override",
		Environment: map[string]string{
			"VAR2": "value2",
		},
		Build: &BuildConfig{
			Dockerfile: "Dockerfile.override",
			Args: map[string]string{
				"ARG2": "arg2",
			},
		},
	}

	base.mergeFrom(override)

	assert.Equal(t, "override", base.Extends)
	assert.Equal(t, "value1", base.Environment["VAR1"])
	assert.Equal(t, "value2", base.Environment["VAR2"])
	assert.Equal(t, "./base", base.Build.Context)
	assert.Equal(t, "Dockerfile.override", base.Build.Dockerfile)
	assert.Equal(t, "arg1", base.Build.Args["ARG1"])
	assert.Equal(t, "arg2", base.Build.Args["ARG2"])
}

func TestBuildConfig_MergeFrom(t *testing.T) {
	base := BuildConfig{
		Context: "./base",
		Args: map[string]string{
			"ARG1": "value1",
		},
	}

	override := BuildConfig{
		Dockerfile: "Dockerfile.override",
		Target:     "production",
		Args: map[string]string{
			"ARG2": "value2",
		},
	}

	base.mergeFrom(override)

	assert.Equal(t, "./base", base.Context)
	assert.Equal(t, "Dockerfile.override", base.Dockerfile)
	assert.Equal(t, "production", base.Target)
	assert.Equal(t, "value1", base.Args["ARG1"])
	assert.Equal(t, "value2", base.Args["ARG2"])
}

func TestLoadDurableComposeConfig_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "durablecompose.toml")

	// Write invalid TOML
	err := os.WriteFile(configPath, []byte("invalid [[[[ toml"), 0644)
	require.NoError(t, err)

	_, err = LoadDurableComposeConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestLoadDurableComposeConfig_FileNotFound(t *testing.T) {
	_, err := LoadDurableComposeConfig("/nonexistent/durablecompose.toml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}
