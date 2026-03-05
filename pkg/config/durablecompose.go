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
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// DurableComposeConfig represents the durablecompose.toml configuration file
type DurableComposeConfig struct {
	// Project-level settings
	Project ProjectConfig `toml:"project,omitempty"`

	// Services configuration
	Services map[string]ServiceConfig `toml:"services,omitempty"`

	// File loading configuration
	Loader LoaderConfig `toml:"loader,omitempty"`

	// Builders configuration (for future extensibility)
	Builders map[string]BuilderConfig `toml:"builders,omitempty"`

	// Raw processes configuration (non-containerized processes)
	Processes ProcessesConfig `toml:"processes,omitempty"`
}

// ProjectConfig contains project-level settings
type ProjectConfig struct {
	// Name of the project
	Name string `toml:"name,omitempty"`

	// Default profiles to enable
	Profiles []string `toml:"profiles,omitempty"`

	// Working directory
	WorkDir string `toml:"work_dir,omitempty"`

	// Environment files to load
	EnvFiles []string `toml:"env_files,omitempty"`
}

// ServiceConfig contains service-specific configuration
type ServiceConfig struct {
	// Extends allows inheriting configuration from another service definition
	Extends string `toml:"extends,omitempty"`

	// ExtendsFile specifies an external file to extend from
	ExtendsFile string `toml:"extends_file,omitempty"`

	// Environment variables to set/override
	Environment map[string]string `toml:"environment,omitempty"`

	// Build configuration overrides
	Build *BuildConfig `toml:"build,omitempty"`

	// Additional compose file paths specific to this service
	ComposeFiles []string `toml:"compose_files,omitempty"`
}

// BuildConfig contains build-specific configuration
type BuildConfig struct {
	// Context directory for build
	Context string `toml:"context,omitempty"`

	// Dockerfile path
	Dockerfile string `toml:"dockerfile,omitempty"`

	// Build arguments
	Args map[string]string `toml:"args,omitempty"`

	// Target stage for multi-stage builds
	Target string `toml:"target,omitempty"`

	// Builder to use (for future extensibility)
	Builder string `toml:"builder,omitempty"`
}

// LoaderConfig configures how compose files are loaded
type LoaderConfig struct {
	// Enabled file types for service loading
	EnabledFileTypes []string `toml:"enabled_file_types,omitempty"`

	// Disabled file types
	DisabledFileTypes []string `toml:"disabled_file_types,omitempty"`

	// Custom file patterns to load
	FilePatterns []string `toml:"file_patterns,omitempty"`

	// Specific files to include
	IncludeFiles []string `toml:"include_files,omitempty"`

	// Specific files to exclude
	ExcludeFiles []string `toml:"exclude_files,omitempty"`

	// Search paths for compose files
	SearchPaths []string `toml:"search_paths,omitempty"`
}

// BuilderConfig represents a custom builder configuration (for future extensibility)
type BuilderConfig struct {
	// Type of builder (docker, buildkit, etc.)
	Type string `toml:"type,omitempty"`

	// Builder-specific options
	Options map[string]interface{} `toml:"options,omitempty"`
}

// ProcessesConfig configures raw process execution
type ProcessesConfig struct {
	// Enabled enables raw process execution
	Enabled bool `toml:"enabled,omitempty"`

	// ProcfilePath specifies the path to a Procfile to load
	ProcfilePath string `toml:"procfile_path,omitempty"`

	// Definitions is a map of process name to process configuration
	Definitions map[string]ProcessDefinition `toml:"definitions,omitempty"`

	// WorkingDir is the default working directory for processes
	WorkingDir string `toml:"working_dir,omitempty"`

	// Environment is a map of default environment variables for all processes
	Environment map[string]string `toml:"environment,omitempty"`
}

// ProcessDefinition defines a single raw process
type ProcessDefinition struct {
	// Command is the shell command to execute
	Command string `toml:"command,omitempty"`

	// WorkingDir overrides the default working directory
	WorkingDir string `toml:"working_dir,omitempty"`

	// Environment overrides/extends default environment variables
	Environment map[string]string `toml:"environment,omitempty"`
}

// LoadDurableComposeConfig loads a durablecompose.toml file from the specified path
func LoadDurableComposeConfig(path string) (*DurableComposeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read durablecompose.toml: %w", err)
	}

	var config DurableComposeConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse durablecompose.toml: %w", err)
	}

	return &config, nil
}

// FindDurableComposeConfig searches for durablecompose.toml in the working directory and parent directories
func FindDurableComposeConfig(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, "durablecompose.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return "", fmt.Errorf("durablecompose.toml not found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

// LoadDurableComposeConfigFromDir searches for and loads durablecompose.toml starting from the specified directory
func LoadDurableComposeConfigFromDir(startDir string) (*DurableComposeConfig, error) {
	configPath, err := FindDurableComposeConfig(startDir)
	if err != nil {
		return nil, err
	}

	return LoadDurableComposeConfig(configPath)
}

// Merge merges another DurableComposeConfig into this one, with other taking precedence
func (c *DurableComposeConfig) Merge(other *DurableComposeConfig) {
	if other.Project.Name != "" {
		c.Project.Name = other.Project.Name
	}
	if len(other.Project.Profiles) > 0 {
		c.Project.Profiles = append(c.Project.Profiles, other.Project.Profiles...)
	}
	if other.Project.WorkDir != "" {
		c.Project.WorkDir = other.Project.WorkDir
	}
	if len(other.Project.EnvFiles) > 0 {
		c.Project.EnvFiles = append(c.Project.EnvFiles, other.Project.EnvFiles...)
	}

	// Merge services
	if c.Services == nil {
		c.Services = make(map[string]ServiceConfig)
	}
	for name, svc := range other.Services {
		if existing, ok := c.Services[name]; ok {
			// Merge service configs
			existing.mergeFrom(svc)
			c.Services[name] = existing
		} else {
			c.Services[name] = svc
		}
	}

	// Merge loader config
	if len(other.Loader.EnabledFileTypes) > 0 {
		c.Loader.EnabledFileTypes = other.Loader.EnabledFileTypes
	}
	if len(other.Loader.DisabledFileTypes) > 0 {
		c.Loader.DisabledFileTypes = other.Loader.DisabledFileTypes
	}
	if len(other.Loader.FilePatterns) > 0 {
		c.Loader.FilePatterns = append(c.Loader.FilePatterns, other.Loader.FilePatterns...)
	}
	if len(other.Loader.IncludeFiles) > 0 {
		c.Loader.IncludeFiles = append(c.Loader.IncludeFiles, other.Loader.IncludeFiles...)
	}
	if len(other.Loader.ExcludeFiles) > 0 {
		c.Loader.ExcludeFiles = append(c.Loader.ExcludeFiles, other.Loader.ExcludeFiles...)
	}
	if len(other.Loader.SearchPaths) > 0 {
		c.Loader.SearchPaths = append(c.Loader.SearchPaths, other.Loader.SearchPaths...)
	}

	// Merge builders
	if c.Builders == nil {
		c.Builders = make(map[string]BuilderConfig)
	}
	for name, builder := range other.Builders {
		c.Builders[name] = builder
	}
}

// mergeFrom merges another ServiceConfig into this one
func (s *ServiceConfig) mergeFrom(other ServiceConfig) {
	if other.Extends != "" {
		s.Extends = other.Extends
	}
	if other.ExtendsFile != "" {
		s.ExtendsFile = other.ExtendsFile
	}
	if s.Environment == nil {
		s.Environment = make(map[string]string)
	}
	for k, v := range other.Environment {
		s.Environment[k] = v
	}
	if other.Build != nil {
		if s.Build == nil {
			s.Build = &BuildConfig{}
		}
		s.Build.mergeFrom(*other.Build)
	}
	if len(other.ComposeFiles) > 0 {
		s.ComposeFiles = append(s.ComposeFiles, other.ComposeFiles...)
	}
}

// mergeFrom merges another BuildConfig into this one
func (b *BuildConfig) mergeFrom(other BuildConfig) {
	if other.Context != "" {
		b.Context = other.Context
	}
	if other.Dockerfile != "" {
		b.Dockerfile = other.Dockerfile
	}
	if b.Args == nil {
		b.Args = make(map[string]string)
	}
	for k, v := range other.Args {
		b.Args[k] = v
	}
	if other.Target != "" {
		b.Target = other.Target
	}
	if other.Builder != "" {
		b.Builder = other.Builder
	}
}
