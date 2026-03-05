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
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// TemplateContext holds runtime context for template interpolation
type TemplateContext struct {
	dockerClient *client.Client
	projectName  string
	serviceLookup map[string]string // maps service name to container ID
}

// NewTemplateContext creates a new template context for interpolation
func NewTemplateContext(dockerClient *client.Client, projectName string) *TemplateContext {
	return &TemplateContext{
		dockerClient: dockerClient,
		projectName:  projectName,
		serviceLookup: make(map[string]string),
	}
}

// RegisterService registers a service name to container ID mapping
func (tc *TemplateContext) RegisterService(serviceName, containerID string) {
	if tc.serviceLookup == nil {
		tc.serviceLookup = make(map[string]string)
	}
	tc.serviceLookup[serviceName] = containerID
}

// GetServiceIP returns the IP address for a given service name
func (tc *TemplateContext) GetServiceIP(ctx context.Context, serviceName string) (string, error) {
	containerID, ok := tc.serviceLookup[serviceName]
	if !ok {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	if tc.dockerClient == nil {
		return "", fmt.Errorf("docker client not initialized")
	}

	// Inspect the container to get its network settings
	containerJSON, err := tc.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	// Try to get the IP from the default network first
	if containerJSON.NetworkSettings != nil && containerJSON.NetworkSettings.IPAddress != "" {
		return containerJSON.NetworkSettings.IPAddress, nil
	}

	// If no IP on default network, try the first available network
	if containerJSON.NetworkSettings != nil && len(containerJSON.NetworkSettings.Networks) > 0 {
		for networkName, network := range containerJSON.NetworkSettings.Networks {
			if network.IPAddress != "" {
				return network.IPAddress, nil
			}
			// If we have a project-specific network, prefer that
			if networkName == tc.projectName+"_default" && network.IPAddress != "" {
				return network.IPAddress, nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found for service %s", serviceName)
}

// InterpolateEnvironment interpolates Jinja2-style templates in environment variable values
func (tc *TemplateContext) InterpolateEnvironment(ctx context.Context, environment map[string]string) (map[string]string, error) {
	if environment == nil {
		return nil, nil
	}

	result := make(map[string]string, len(environment))

	// Create template context with service_ip function
	templateCtx := exec.NewContext(map[string]interface{}{
		"service_ip": func(serviceName string) (string, error) {
			return tc.GetServiceIP(ctx, serviceName)
		},
	})

	for key, value := range environment {
		// Check if the value contains template syntax
		if !containsTemplate(value) {
			result[key] = value
			continue
		}

		// Parse and render the template
		tpl, err := gonja.FromString(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template for %s: %w", key, err)
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, templateCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to render template for %s: %w", key, err)
		}

		result[key] = buf.String()
	}

	return result, nil
}

// containsTemplate checks if a string contains Jinja2 template syntax
func containsTemplate(s string) bool {
	return len(s) >= 4 && (hasSequence(s, "{{") || hasSequence(s, "{%"))
}

// hasSequence checks if a string contains a specific sequence
func hasSequence(s, seq string) bool {
	for i := 0; i <= len(s)-len(seq); i++ {
		if s[i:i+len(seq)] == seq {
			return true
		}
	}
	return false
}

// InterpolateServiceEnvironment interpolates environment variables for a specific service
func (c *DurableComposeConfig) InterpolateServiceEnvironment(ctx context.Context, serviceName string, tc *TemplateContext) (map[string]string, error) {
	env := c.GetServiceEnvironment(serviceName)
	if env == nil {
		return nil, nil
	}

	return tc.InterpolateEnvironment(ctx, env)
}
