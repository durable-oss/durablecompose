/*
   Copyright 2023 Docker Compose CLI authors

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

package oci

import (
	"testing"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
)

func TestNewResolver(t *testing.T) {
	tests := []struct {
		name                string
		config              *configfile.ConfigFile
		insecureRegistries  []string
		expectNonNil        bool
	}{
		{
			name: "empty config",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{},
			},
			insecureRegistries: nil,
			expectNonNil:       true,
		},
		{
			name: "config with auth",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{
					"docker.io": {
						Username: "testuser",
						Password: "testpass",
					},
				},
			},
			insecureRegistries: nil,
			expectNonNil:       true,
		},
		{
			name: "config with identity token",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{
					"gcr.io": {
						IdentityToken: "token123",
					},
				},
			},
			insecureRegistries: nil,
			expectNonNil:       true,
		},
		{
			name: "with insecure registries",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{},
			},
			insecureRegistries: []string{"localhost:5000", "registry.local"},
			expectNonNil:       true,
		},
		{
			name: "nil config",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{},
			},
			insecureRegistries: nil,
			expectNonNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.config, tt.insecureRegistries...)
			if tt.expectNonNil && resolver == nil {
				t.Error("NewResolver() returned nil, expected non-nil resolver")
			}
		})
	}
}

func TestNewResolverWithMultipleInsecureRegistries(t *testing.T) {
	config := &configfile.ConfigFile{
		AuthConfigs: map[string]types.AuthConfig{},
	}
	insecureRegistries := []string{
		"localhost:5000",
		"registry1.local",
		"registry2.local:8080",
	}

	resolver := NewResolver(config, insecureRegistries...)
	if resolver == nil {
		t.Fatal("NewResolver() returned nil")
	}
}

func TestNewResolverEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *configfile.ConfigFile
	}{
		{
			name: "empty auth configs",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{},
			},
		},
		{
			name: "multiple auth configs",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{
					"docker.io": {
						Username: "user1",
						Password: "pass1",
					},
					"gcr.io": {
						IdentityToken: "token1",
					},
					"ghcr.io": {
						Username: "user2",
						Password: "pass2",
					},
				},
			},
		},
		{
			name: "auth config with empty credentials",
			config: &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{
					"docker.io": {
						Username: "",
						Password: "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.config)
			if resolver == nil {
				t.Error("NewResolver() returned nil")
			}
		})
	}
}

func TestNewResolverAuthConfigVariations(t *testing.T) {
	tests := []struct {
		name       string
		authConfig types.AuthConfig
	}{
		{
			name: "username and password",
			authConfig: types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
		},
		{
			name: "identity token only",
			authConfig: types.AuthConfig{
				IdentityToken: "token123",
			},
		},
		{
			name: "username password and identity token",
			authConfig: types.AuthConfig{
				Username:      "testuser",
				Password:      "testpass",
				IdentityToken: "token123",
			},
		},
		{
			name: "empty auth config",
			authConfig: types.AuthConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configfile.ConfigFile{
				AuthConfigs: map[string]types.AuthConfig{
					"test.registry.io": tt.authConfig,
				},
			}
			resolver := NewResolver(config)
			if resolver == nil {
				t.Error("NewResolver() returned nil")
			}
		})
	}
}
