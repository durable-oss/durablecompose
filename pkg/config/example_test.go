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

package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/durable_oss/durablecompose/pkg/config"
)

func ExampleLoadDurableComposeConfig() {
	// Create a temporary config file for the example
	tmpDir, _ := os.MkdirTemp("", "example")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "durablecompose.toml")
	configContent := `
[project]
name = "example-app"
profiles = ["dev", "test"]

[services.web]
[services.web.environment]
PORT = "8080"
DEBUG = "true"

[loader]
enabled_file_types = ["yml", "yaml"]
`
	_ = os.WriteFile(configPath, []byte(configContent), 0644)

	// Load the configuration
	cfg, _ := config.LoadDurableComposeConfig(configPath)

	fmt.Println("Project Name:", cfg.Project.Name)
	fmt.Println("Profiles:", cfg.Project.Profiles)
	fmt.Println("Web Port:", cfg.GetServiceEnvironment("web")["PORT"])
	fmt.Println("Enabled Types:", cfg.Loader.EnabledFileTypes)

	// Output:
	// Project Name: example-app
	// Profiles: [dev test]
	// Web Port: 8080
	// Enabled Types: [yml yaml]
}

func ExampleDurableComposeConfig_FilterComposeFiles() {
	cfg := &config.DurableComposeConfig{
		Loader: config.LoaderConfig{
			EnabledFileTypes: []string{"yml"},
			ExcludeFiles:     []string{"*.override.yml"},
		},
	}

	files := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"docker-compose.override.yml",
	}

	filtered := cfg.FilterComposeFiles(files)
	fmt.Println(filtered)

	// Output:
	// [docker-compose.yml]
}

func ExampleDurableComposeConfig_GetSearchPaths() {
	cfg := &config.DurableComposeConfig{
		Loader: config.LoaderConfig{
			SearchPaths: []string{"./services", "/opt/configs"},
		},
	}

	paths := cfg.GetSearchPaths("/workspace")
	for _, path := range paths {
		fmt.Println(path)
	}

	// Output:
	// /workspace
	// /workspace/services
	// /opt/configs
}
