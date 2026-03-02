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

package registry

import (
	"testing"
)

func TestGetAuthConfigKey(t *testing.T) {
	tests := []struct {
		name      string
		indexName string
		want      string
	}{
		{
			name:      "docker.io index",
			indexName: IndexName,
			want:      IndexServer,
		},
		{
			name:      "index.docker.io hostname",
			indexName: IndexHostname,
			want:      IndexServer,
		},
		{
			name:      "registry-1.docker.io hostname",
			indexName: DefaultRegistryHost,
			want:      IndexServer,
		},
		{
			name:      "private registry with port",
			indexName: "registry.example.com:5000",
			want:      "registry.example.com:5000",
		},
		{
			name:      "private registry without port",
			indexName: "registry.example.com",
			want:      "registry.example.com",
		},
		{
			name:      "gcr.io registry",
			indexName: "gcr.io",
			want:      "gcr.io",
		},
		{
			name:      "ghcr.io registry",
			indexName: "ghcr.io",
			want:      "ghcr.io",
		},
		{
			name:      "localhost registry",
			indexName: "localhost:5000",
			want:      "localhost:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAuthConfigKey(tt.indexName)
			if got != tt.want {
				t.Errorf("GetAuthConfigKey(%q) = %q, want %q", tt.indexName, got, tt.want)
			}
		})
	}
}

func TestGetAuthConfigKeyEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		indexName string
		want      string
	}{
		{
			name:      "empty string",
			indexName: "",
			want:      "",
		},
		{
			name:      "IP address with port",
			indexName: "192.168.1.100:5000",
			want:      "192.168.1.100:5000",
		},
		{
			name:      "IP address without port",
			indexName: "192.168.1.100",
			want:      "192.168.1.100",
		},
		{
			name:      "registry with subdomain",
			indexName: "eu.gcr.io",
			want:      "eu.gcr.io",
		},
		{
			name:      "case sensitive registry name",
			indexName: "Registry.Example.Com",
			want:      "Registry.Example.Com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAuthConfigKey(tt.indexName)
			if got != tt.want {
				t.Errorf("GetAuthConfigKey(%q) = %q, want %q", tt.indexName, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify that constants have expected values
	if DefaultNamespace != "docker.io" {
		t.Errorf("DefaultNamespace = %q, want %q", DefaultNamespace, "docker.io")
	}

	if DefaultRegistryHost != "registry-1.docker.io" {
		t.Errorf("DefaultRegistryHost = %q, want %q", DefaultRegistryHost, "registry-1.docker.io")
	}

	if IndexHostname != "index.docker.io" {
		t.Errorf("IndexHostname = %q, want %q", IndexHostname, "index.docker.io")
	}

	if IndexName != "docker.io" {
		t.Errorf("IndexName = %q, want %q", IndexName, "docker.io")
	}

	expectedIndexServer := "https://index.docker.io/v1/"
	if IndexServer != expectedIndexServer {
		t.Errorf("IndexServer = %q, want %q", IndexServer, expectedIndexServer)
	}
}
