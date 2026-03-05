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

	"github.com/compose-spec/compose-go/v2/cli"
)

// ApplyToProjectOptions applies durablecompose.toml configuration to compose-go ProjectOptions
func (c *DurableComposeConfig) ApplyToProjectOptions(opts []cli.ProjectOptionsFn) []cli.ProjectOptionsFn {
	if c == nil {
		return opts
	}

	// Apply project name
	if c.Project.Name != "" {
		opts = append(opts, cli.WithName(c.Project.Name))
	}

	// Apply profiles
	if len(c.Project.Profiles) > 0 {
		opts = append(opts, cli.WithDefaultProfiles(c.Project.Profiles...))
	}

	// Apply working directory
	if c.Project.WorkDir != "" {
		opts = append(opts, cli.WithWorkingDirectory(c.Project.WorkDir))
	}

	// Apply env files
	if len(c.Project.EnvFiles) > 0 {
		opts = append(opts, cli.WithEnvFiles(c.Project.EnvFiles...))
	}

	// Apply loader config - customize file patterns
	if len(c.Loader.IncludeFiles) > 0 {
		// Add specific files to config paths if they exist
		for _, file := range c.Loader.IncludeFiles {
			if _, err := os.Stat(file); err == nil {
				opts = append(opts, cli.WithConfigFileEnv)
			}
		}
	}

	return opts
}

// LoadAndApplyConfig loads durablecompose.toml from the given directory and applies it
func LoadAndApplyConfig(workDir string, opts []cli.ProjectOptionsFn) ([]cli.ProjectOptionsFn, *DurableComposeConfig, error) {
	// Try to find and load durablecompose.toml
	config, err := LoadDurableComposeConfigFromDir(workDir)
	if err != nil {
		// Config file not found is not an error - it's optional
		return opts, nil, nil
	}

	// Apply the configuration
	opts = config.ApplyToProjectOptions(opts)

	return opts, config, nil
}

// FilterComposeFiles filters compose file paths based on loader configuration
func (c *DurableComposeConfig) FilterComposeFiles(files []string) []string {
	if c == nil || (len(c.Loader.IncludeFiles) == 0 && len(c.Loader.ExcludeFiles) == 0 && len(c.Loader.FilePatterns) == 0 && len(c.Loader.EnabledFileTypes) == 0 && len(c.Loader.DisabledFileTypes) == 0) {
		return files
	}

	filtered := make([]string, 0, len(files))

	for _, file := range files {
		if c.shouldIncludeFile(file) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// shouldIncludeFile checks if a file should be included based on loader configuration
func (c *DurableComposeConfig) shouldIncludeFile(file string) bool {
	basename := filepath.Base(file)

	// Check explicit exclusions first
	for _, exclude := range c.Loader.ExcludeFiles {
		if match, _ := filepath.Match(exclude, basename); match {
			return false
		}
	}

	// Check explicit inclusions
	if len(c.Loader.IncludeFiles) > 0 {
		for _, include := range c.Loader.IncludeFiles {
			if match, _ := filepath.Match(include, basename); match {
				return true
			}
		}
		return false
	}

	// Check file patterns
	if len(c.Loader.FilePatterns) > 0 {
		for _, pattern := range c.Loader.FilePatterns {
			if match, _ := filepath.Match(pattern, basename); match {
				return true
			}
		}
		return false
	}

	// Check file type filters
	if len(c.Loader.EnabledFileTypes) > 0 || len(c.Loader.DisabledFileTypes) > 0 {
		ext := filepath.Ext(file)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:] // Remove leading dot
		}

		// Check disabled types first
		for _, disabled := range c.Loader.DisabledFileTypes {
			if ext == disabled {
				return false
			}
		}

		// Check enabled types
		if len(c.Loader.EnabledFileTypes) > 0 {
			for _, enabled := range c.Loader.EnabledFileTypes {
				if ext == enabled {
					return true
				}
			}
			return false
		}

		// If we only had disabled types and didn't match any, include the file
		return true
	}

	// Default: include the file
	return true
}

// GetSearchPaths returns all search paths for compose files
func (c *DurableComposeConfig) GetSearchPaths(workDir string) []string {
	if c == nil || len(c.Loader.SearchPaths) == 0 {
		return []string{workDir}
	}

	paths := make([]string, 0, len(c.Loader.SearchPaths)+1)
	paths = append(paths, workDir)

	for _, path := range c.Loader.SearchPaths {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
		} else {
			paths = append(paths, filepath.Join(workDir, path))
		}
	}

	return paths
}

// GetServiceEnvironment returns environment variables for a specific service
func (c *DurableComposeConfig) GetServiceEnvironment(serviceName string) map[string]string {
	if c == nil || c.Services == nil {
		return nil
	}

	if svc, ok := c.Services[serviceName]; ok {
		return svc.Environment
	}

	return nil
}

// GetServiceBuildConfig returns build configuration for a specific service
func (c *DurableComposeConfig) GetServiceBuildConfig(serviceName string) *BuildConfig {
	if c == nil || c.Services == nil {
		return nil
	}

	if svc, ok := c.Services[serviceName]; ok {
		return svc.Build
	}

	return nil
}
