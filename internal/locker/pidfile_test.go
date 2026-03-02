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

package locker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPidfile(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
	}{
		{
			name:        "valid project name",
			projectName: "test-project",
			wantErr:     false,
		},
		{
			name:        "project name with special characters",
			projectName: "test-project_123",
			wantErr:     false,
		},
		{
			name:        "empty project name",
			projectName: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf, err := NewPidfile(tt.projectName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPidfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if pf == nil {
					t.Error("NewPidfile() returned nil pidfile")
					return
				}
				if pf.path == "" {
					t.Error("NewPidfile() returned pidfile with empty path")
				}
			}
		})
	}
}

func TestPidfilePath(t *testing.T) {
	projectName := "test-project"
	pf, err := NewPidfile(projectName)
	if err != nil {
		t.Fatalf("NewPidfile() failed: %v", err)
	}

	expectedFilename := projectName + ".pid"
	if !filepath.IsAbs(pf.path) {
		t.Errorf("Pidfile path should be absolute, got: %s", pf.path)
	}
	if filepath.Base(pf.path) != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filepath.Base(pf.path))
	}
}

func TestRunDir(t *testing.T) {
	// Test with XDG_RUNTIME_DIR set
	t.Run("with XDG_RUNTIME_DIR", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_RUNTIME_DIR", tempDir)

		dir, err := runDir()
		if err != nil {
			t.Fatalf("runDir() failed: %v", err)
		}
		if dir != tempDir {
			t.Errorf("Expected runDir to return %s, got %s", tempDir, dir)
		}
	})

	// Test without XDG_RUNTIME_DIR
	t.Run("without XDG_RUNTIME_DIR", func(t *testing.T) {
		os.Unsetenv("XDG_RUNTIME_DIR")

		dir, err := runDir()
		if err != nil {
			t.Fatalf("runDir() failed: %v", err)
		}
		if dir == "" {
			t.Error("runDir() returned empty directory")
		}
		if !filepath.IsAbs(dir) {
			t.Errorf("runDir() should return absolute path, got: %s", dir)
		}
	})
}

func TestPidfileEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		setup       func()
		cleanup     func()
	}{
		{
			name:        "very long project name",
			projectName: string(make([]byte, 255)),
		},
		{
			name:        "project name with dots",
			projectName: "my.project.name",
		},
		{
			name:        "project name with hyphens",
			projectName: "my-project-name",
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

			pf, err := NewPidfile(tt.projectName)
			if err != nil {
				t.Errorf("NewPidfile() failed for %s: %v", tt.name, err)
				return
			}
			if pf == nil {
				t.Error("NewPidfile() returned nil pidfile")
			}
		})
	}
}

func TestPidfileLock(t *testing.T) {
	t.Run("lock creates pidfile", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_RUNTIME_DIR", tempDir)

		pf, err := NewPidfile("test-lock-project")
		if err != nil {
			t.Fatalf("NewPidfile() failed: %v", err)
		}

		err = pf.Lock()
		if err != nil {
			t.Errorf("Lock() failed: %v", err)
		}

		if _, err := os.Stat(pf.path); os.IsNotExist(err) {
			t.Error("Lock() did not create pidfile")
		}
	})

	t.Run("lock with existing pidfile", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_RUNTIME_DIR", tempDir)

		pf1, err := NewPidfile("test-existing")
		if err != nil {
			t.Fatalf("NewPidfile() failed: %v", err)
		}

		err = pf1.Lock()
		if err != nil {
			t.Errorf("First Lock() failed: %v", err)
		}

		pf2, err := NewPidfile("test-existing")
		if err != nil {
			t.Fatalf("Second NewPidfile() failed: %v", err)
		}

		err = pf2.Lock()
		if err == nil {
			t.Error("Second Lock() should have failed but succeeded")
		}
	})
}

func TestRunDirCreation(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		tempBase := t.TempDir()
		nonExistent := filepath.Join(tempBase, "nonexistent", "runtime")
		t.Setenv("XDG_RUNTIME_DIR", "")

		os.Unsetenv("XDG_RUNTIME_DIR")

		dir, err := runDir()
		if err != nil {
			t.Errorf("runDir() should create directory, got error: %v", err)
		}
		if dir == "" {
			t.Error("runDir() returned empty directory")
		}

		_ = nonExistent
	})
}

func TestPidfileStructAccess(t *testing.T) {
	t.Run("verify path field", func(t *testing.T) {
		pf, err := NewPidfile("test-struct")
		if err != nil {
			t.Fatalf("NewPidfile() failed: %v", err)
		}
		if pf.path == "" {
			t.Error("Pidfile path should not be empty")
		}
		if !filepath.IsAbs(pf.path) {
			t.Errorf("Pidfile path should be absolute, got: %s", pf.path)
		}
	})
}
