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

package procfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestParse_Defensive(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := Parse("")
		assert.ErrorIs(t, err, ErrEmptyPath)
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := Parse("/nonexistent/path/to/Procfile")
		assert.ErrorContains(t, err, "does not exist")
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := Parse(tmpDir)
		assert.ErrorContains(t, err, "not a regular file")
	})

	t.Run("invalid process name - too long", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		longName := strings.Repeat("a", MaxProcessNameLength+1)
		content := longName + ": echo test"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorIs(t, err, ErrProcessNameTooLong)
	})

	t.Run("invalid process name - starts with hyphen", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		content := "-web: echo test"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "cannot start with hyphen")
	})

	t.Run("invalid process name - starts with underscore", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		content := "_web: echo test"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "cannot start with")
	})

	t.Run("invalid process name - contains invalid characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		invalidNames := []string{
			"web@server: echo test",
			"web server: echo test",
			"web.server: echo test",
			"web$var: echo test",
		}

		for _, invalidContent := range invalidNames {
			err := os.WriteFile(procfilePath, []byte(invalidContent), 0644)
			assert.NilError(t, err)

			_, err = Parse(procfilePath)
			assert.ErrorIs(t, err, ErrInvalidProcessName, "expected error for: %s", invalidContent)
		}
	})

	t.Run("invalid process name - contains control characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		content := "web\x00: echo test"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "invalid process name")
	})

	t.Run("command too long", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		longCommand := strings.Repeat("a", MaxCommandLength+1)
		content := "web: " + longCommand
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorIs(t, err, ErrCommandTooLong)
	})

	t.Run("command with null byte", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		content := "web: echo test\x00"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "null byte")
	})

	t.Run("too many processes", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		var content strings.Builder
		for i := 0; i <= MaxProcesses; i++ {
			// Create unique names to avoid duplicates
			content.WriteString(fmt.Sprintf("proc%d: echo test\n", i))
		}

		err := os.WriteFile(procfilePath, []byte(content.String()), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorIs(t, err, ErrTooManyProcesses)
	})

	t.Run("duplicate names - case insensitive", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		content := "web: echo test1\nWeb: echo test2"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "duplicate process name")
	})

	t.Run("file too large", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		// Create a file larger than 10MB
		largeContent := strings.Repeat("web: echo test\n", 1024*1024)
		err := os.WriteFile(procfilePath, []byte(largeContent), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		assert.ErrorContains(t, err, "too large")
	})

	t.Run("line too long", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		longLine := "web: " + strings.Repeat("a", MaxLineLength)
		err := os.WriteFile(procfilePath, []byte(longLine), 0644)
		assert.NilError(t, err)

		_, err = Parse(procfilePath)
		// Scanner will error with "token too long" before we can check
		assert.ErrorContains(t, err, "token too long")
	})

	t.Run("valid edge cases", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		// Maximum valid name length
		maxName := strings.Repeat("a", MaxProcessNameLength)
		content := maxName + ": echo test"
		err := os.WriteFile(procfilePath, []byte(content), 0644)
		assert.NilError(t, err)

		pf, err := Parse(procfilePath)
		assert.NilError(t, err)
		assert.Equal(t, 1, len(pf.Processes))
	})
}

func TestFindProcfile_Defensive(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := FindProcfile("")
		assert.ErrorIs(t, err, ErrEmptyPath)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := FindProcfile("/nonexistent/directory")
		assert.ErrorContains(t, err, "does not exist")
	})

	t.Run("procfile is a directory not a file", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a directory named "Procfile"
		procfileDir := filepath.Join(tmpDir, "Procfile")
		err := os.Mkdir(procfileDir, 0755)
		assert.NilError(t, err)

		// Should not find it (should continue searching)
		_, err = FindProcfile(tmpDir)
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("finds procfile in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "sub1", "sub2", "sub3")
		err := os.MkdirAll(subDir, 0755)
		assert.NilError(t, err)

		// Create Procfile in root
		procfilePath := filepath.Join(tmpDir, "Procfile")
		err = os.WriteFile(procfilePath, []byte("web: echo test"), 0644)
		assert.NilError(t, err)

		// Search from deep subdirectory
		found, err := FindProcfile(subDir)
		assert.NilError(t, err)
		assert.Equal(t, procfilePath, found)
	})
}

func TestParseFromDir_Defensive(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := ParseFromDir("")
		assert.ErrorIs(t, err, ErrEmptyPath)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := ParseFromDir("/nonexistent")
		assert.ErrorContains(t, err, "failed to find Procfile")
	})

	t.Run("procfile found but invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		procfilePath := filepath.Join(tmpDir, "Procfile")

		// Create invalid Procfile
		err := os.WriteFile(procfilePath, []byte("invalid content without colon"), 0644)
		assert.NilError(t, err)

		_, err = ParseFromDir(tmpDir)
		assert.ErrorContains(t, err, "failed to parse Procfile")
	})
}

func TestValidate(t *testing.T) {
	t.Run("nil procfile", func(t *testing.T) {
		var pf *Procfile
		err := pf.Validate()
		assert.ErrorIs(t, err, ErrNilProcfile)
	})

	t.Run("empty procfile", func(t *testing.T) {
		pf := &Procfile{
			Processes: make(map[string]Process),
		}
		err := pf.Validate()
		assert.ErrorContains(t, err, "contains no processes")
	})

	t.Run("process name mismatch", func(t *testing.T) {
		pf := &Procfile{
			Processes: map[string]Process{
				"web": {
					Name:    "worker",
					Command: "echo test",
				},
			},
		}
		err := pf.Validate()
		assert.ErrorContains(t, err, "mismatch")
	})

	t.Run("valid procfile", func(t *testing.T) {
		pf := &Procfile{
			Processes: map[string]Process{
				"web": {
					Name:    "web",
					Command: "echo test",
				},
			},
		}
		err := pf.Validate()
		assert.NilError(t, err)
	})
}

func TestGetProcess(t *testing.T) {
	t.Run("nil procfile", func(t *testing.T) {
		var pf *Procfile
		_, exists := pf.GetProcess("web")
		assert.Assert(t, !exists)
	})

	t.Run("process exists", func(t *testing.T) {
		pf := &Procfile{
			Processes: map[string]Process{
				"web": {
					Name:    "web",
					Command: "echo test",
				},
			},
		}
		proc, exists := pf.GetProcess("web")
		assert.Assert(t, exists)
		assert.Equal(t, "web", proc.Name)
	})

	t.Run("process does not exist", func(t *testing.T) {
		pf := &Procfile{
			Processes: map[string]Process{
				"web": {
					Name:    "web",
					Command: "echo test",
				},
			},
		}
		_, exists := pf.GetProcess("worker")
		assert.Assert(t, !exists)
	})
}

func TestProcessNames(t *testing.T) {
	t.Run("nil procfile", func(t *testing.T) {
		var pf *Procfile
		names := pf.ProcessNames()
		assert.Assert(t, names == nil)
	})

	t.Run("empty procfile", func(t *testing.T) {
		pf := &Procfile{
			Processes: make(map[string]Process),
		}
		names := pf.ProcessNames()
		assert.Assert(t, names == nil)
	})

	t.Run("with processes", func(t *testing.T) {
		pf := &Procfile{
			Processes: map[string]Process{
				"web":    {Name: "web", Command: "echo 1"},
				"worker": {Name: "worker", Command: "echo 2"},
			},
		}
		names := pf.ProcessNames()
		assert.Equal(t, 2, len(names))
		assert.Assert(t, contains(names, "web"))
		assert.Assert(t, contains(names, "worker"))
	})
}

func TestValidateProcessName(t *testing.T) {
	tests := []struct {
		name      string
		procName  string
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name:     "valid name",
			procName: "web",
			wantErr:  false,
		},
		{
			name:     "valid with numbers",
			procName: "web123",
			wantErr:  false,
		},
		{
			name:     "valid with hyphen",
			procName: "web-server",
			wantErr:  false,
		},
		{
			name:     "valid with underscore",
			procName: "web_server",
			wantErr:  false,
		},
		{
			name:     "empty",
			procName: "",
			wantErr:  true,
		},
		{
			name:     "too long",
			procName: strings.Repeat("a", MaxProcessNameLength+1),
			wantErr:  true,
			errCheck: func(err error) bool { return errors.Is(err, ErrProcessNameTooLong) },
		},
		{
			name:     "starts with hyphen",
			procName: "-web",
			wantErr:  true,
		},
		{
			name:     "starts with underscore",
			procName: "_web",
			wantErr:  true,
		},
		{
			name:     "contains space",
			procName: "web server",
			wantErr:  true,
			errCheck: func(err error) bool { return errors.Is(err, ErrInvalidProcessName) },
		},
		{
			name:     "contains special char",
			procName: "web@server",
			wantErr:  true,
			errCheck: func(err error) bool { return errors.Is(err, ErrInvalidProcessName) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProcessName(tt.procName)
			if tt.wantErr {
				assert.Assert(t, err != nil, "expected error for %q", tt.procName)
				if tt.errCheck != nil {
					assert.Assert(t, tt.errCheck(err), "error check failed for %q: %v", tt.procName, err)
				}
			} else {
				assert.NilError(t, err, "unexpected error for %q", tt.procName)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "valid command",
			command: "npm start",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:     "command too long",
			command:  strings.Repeat("a", MaxCommandLength+1),
			wantErr:  true,
			errCheck: func(err error) bool { return errors.Is(err, ErrCommandTooLong) },
		},
		{
			name:    "command with null byte",
			command: "echo test\x00",
			wantErr: true,
		},
		{
			name:    "complex valid command",
			command: `sh -c "echo 'test' && npm start"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.command)
			if tt.wantErr {
				assert.Assert(t, err != nil, "expected error for %q", tt.command)
				if tt.errCheck != nil {
					assert.Assert(t, tt.errCheck(err), "error check failed")
				}
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestSanitizeForError(t *testing.T) {
	t.Run("short string", func(t *testing.T) {
		result := sanitizeForError("hello")
		assert.Equal(t, "hello", result)
	})

	t.Run("long string truncated", func(t *testing.T) {
		long := strings.Repeat("a", 150)
		result := sanitizeForError(long)
		assert.Assert(t, len(result) <= 104) // 100 + "..."
		assert.Assert(t, strings.HasSuffix(result, "..."))
	})

	t.Run("control characters replaced", func(t *testing.T) {
		input := "hello\x00\x01world"
		result := sanitizeForError(input)
		assert.Assert(t, !strings.Contains(result, "\x00"))
		assert.Assert(t, !strings.Contains(result, "\x01"))
	})
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
