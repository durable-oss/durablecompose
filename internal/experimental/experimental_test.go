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

package experimental

import (
	"context"
	"os"
	"testing"

	"github.com/durable_oss/durablecompose/internal/desktop"
)

func TestNewState(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    bool
		setup   func()
		cleanup func()
	}{
		{
			name:   "COMPOSE_EXPERIMENTAL set to 1",
			envVal: "1",
			want:   true,
			setup: func() {
				os.Setenv("COMPOSE_EXPERIMENTAL", "1")
			},
			cleanup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
		},
		{
			name:   "COMPOSE_EXPERIMENTAL set to true",
			envVal: "true",
			want:   true,
			setup: func() {
				os.Setenv("COMPOSE_EXPERIMENTAL", "true")
			},
			cleanup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
		},
		{
			name:   "COMPOSE_EXPERIMENTAL set to 0",
			envVal: "0",
			want:   false,
			setup: func() {
				os.Setenv("COMPOSE_EXPERIMENTAL", "0")
			},
			cleanup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
		},
		{
			name:   "COMPOSE_EXPERIMENTAL not set",
			envVal: "",
			want:   true, // Default is true when not set
			setup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
			cleanup: func() {},
		},
		{
			name:   "COMPOSE_EXPERIMENTAL set to empty string",
			envVal: "",
			want:   true, // Empty string means default (true)
			setup: func() {
				os.Setenv("COMPOSE_EXPERIMENTAL", "")
			},
			cleanup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
		},
		{
			name:   "COMPOSE_EXPERIMENTAL set to false",
			envVal: "false",
			want:   false,
			setup: func() {
				os.Setenv("COMPOSE_EXPERIMENTAL", "false")
			},
			cleanup: func() {
				os.Unsetenv("COMPOSE_EXPERIMENTAL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			state := NewState()
			if state.active != tt.want {
				t.Errorf("NewState().active = %v, want %v", state.active, tt.want)
			}
		})
	}
}

func TestNewStateEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{
			name:   "uppercase TRUE",
			envVal: "TRUE",
			want:   true,
		},
		{
			name:   "mixed case True",
			envVal: "True",
			want:   true,
		},
		{
			name:   "invalid value",
			envVal: "invalid",
			want:   false,
		},
		{
			name:   "numeric 2",
			envVal: "2",
			want:   false,
		},
		{
			name:   "yes",
			envVal: "yes",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("COMPOSE_EXPERIMENTAL", tt.envVal)
			defer os.Unsetenv("COMPOSE_EXPERIMENTAL")

			state := NewState()
			if state.active != tt.want {
				t.Errorf("NewState().active with %q = %v, want %v", tt.envVal, state.active, tt.want)
			}
		})
	}
}

func TestStateLoad(t *testing.T) {
	tests := []struct {
		name        string
		setupState  func() *State
		client      *desktop.Client
		expectError bool
	}{
		{
			name: "inactive state - no load needed",
			setupState: func() *State {
				return &State{active: false}
			},
			client:      nil,
			expectError: false,
		},
		{
			name: "active state - nil client",
			setupState: func() *State {
				return &State{active: true}
			},
			client:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setupState()
			err := state.Load(context.Background(), tt.client)
			if (err != nil) != tt.expectError {
				t.Errorf("Load() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateActiveField(t *testing.T) {
	t.Run("verify active field accessibility", func(t *testing.T) {
		os.Setenv("COMPOSE_EXPERIMENTAL", "1")
		defer os.Unsetenv("COMPOSE_EXPERIMENTAL")

		state := NewState()
		if !state.active {
			t.Error("expected active to be true when COMPOSE_EXPERIMENTAL=1")
		}

		os.Setenv("COMPOSE_EXPERIMENTAL", "0")
		state2 := NewState()
		if state2.active {
			t.Error("expected active to be false when COMPOSE_EXPERIMENTAL=0")
		}
	})
}
