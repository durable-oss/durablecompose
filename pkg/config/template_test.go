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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainsTemplate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "no template",
			input:    "plain text",
			expected: false,
		},
		{
			name:     "variable template",
			input:    "{{ variable }}",
			expected: true,
		},
		{
			name:     "statement template",
			input:    "{% if condition %}",
			expected: true,
		},
		{
			name:     "mixed content",
			input:    "prefix {{ variable }} suffix",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single brace",
			input:    "{text}",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsTemplate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterpolateEnvironment_NoTemplates(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	environment := map[string]string{
		"PLAIN_VAR": "plain_value",
		"ANOTHER":   "another_value",
	}

	result, err := tc.InterpolateEnvironment(ctx, environment)
	require.NoError(t, err)
	assert.Equal(t, environment, result)
}

func TestInterpolateEnvironment_SimpleTemplate(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	environment := map[string]string{
		"GREETING": "{{ 'Hello, World!' }}",
		"PLAIN":    "no template",
	}

	result, err := tc.InterpolateEnvironment(ctx, environment)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result["GREETING"])
	assert.Equal(t, "no template", result["PLAIN"])
}

func TestInterpolateEnvironment_InvalidTemplate(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	environment := map[string]string{
		"INVALID": "{{ unclosed",
	}

	_, err := tc.InterpolateEnvironment(ctx, environment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestInterpolateEnvironment_NilEnvironment(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	result, err := tc.InterpolateEnvironment(ctx, nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRegisterService(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")

	tc.RegisterService("web", "container123")
	tc.RegisterService("db", "container456")

	assert.Equal(t, "container123", tc.serviceLookup["web"])
	assert.Equal(t, "container456", tc.serviceLookup["db"])
}

func TestGetServiceIP_ServiceNotFound(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	_, err := tc.GetServiceIP(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service nonexistent not found")
}

func TestGetServiceIP_NoDockerClient(t *testing.T) {
	tc := NewTemplateContext(nil, "test-project")
	tc.RegisterService("web", "container123")
	ctx := context.Background()

	_, err := tc.GetServiceIP(ctx, "web")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "docker client not initialized")
}

func TestHasSequence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sequence string
		expected bool
	}{
		{
			name:     "sequence present",
			input:    "hello {{world}}",
			sequence: "{{",
			expected: true,
		},
		{
			name:     "sequence not present",
			input:    "hello world",
			sequence: "{{",
			expected: false,
		},
		{
			name:     "sequence at start",
			input:    "{{hello}}",
			sequence: "{{",
			expected: true,
		},
		{
			name:     "sequence at end",
			input:    "hello{{",
			sequence: "{{",
			expected: true,
		},
		{
			name:     "empty input",
			input:    "",
			sequence: "{{",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSequence(tt.input, tt.sequence)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterpolateServiceEnvironment(t *testing.T) {
	config := &DurableComposeConfig{
		Services: map[string]ServiceConfig{
			"web": {
				Environment: map[string]string{
					"STATIC": "value",
					"TEMPLATE": "{{ 'rendered' }}",
				},
			},
		},
	}

	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	result, err := config.InterpolateServiceEnvironment(ctx, "web", tc)
	require.NoError(t, err)
	assert.Equal(t, "value", result["STATIC"])
	assert.Equal(t, "rendered", result["TEMPLATE"])
}

func TestInterpolateServiceEnvironment_ServiceNotFound(t *testing.T) {
	config := &DurableComposeConfig{
		Services: map[string]ServiceConfig{},
	}

	tc := NewTemplateContext(nil, "test-project")
	ctx := context.Background()

	result, err := config.InterpolateServiceEnvironment(ctx, "nonexistent", tc)
	require.NoError(t, err)
	assert.Nil(t, result)
}
