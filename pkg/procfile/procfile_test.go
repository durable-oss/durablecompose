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
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		expected    map[string]string
	}{
		{
			name: "valid procfile",
			content: `web: npm start
worker: node worker.js
scheduler: python scheduler.py`,
			wantErr: false,
			expected: map[string]string{
				"web":       "npm start",
				"worker":    "node worker.js",
				"scheduler": "python scheduler.py",
			},
		},
		{
			name: "with comments and empty lines",
			content: `# This is a comment
web: npm start

# Another comment
worker: node worker.js
`,
			wantErr: false,
			expected: map[string]string{
				"web":    "npm start",
				"worker": "node worker.js",
			},
		},
		{
			name: "command with colons",
			content: `web: sh -c "echo 'Starting server on port:8080'"
`,
			wantErr: false,
			expected: map[string]string{
				"web": `sh -c "echo 'Starting server on port:8080'"`,
			},
		},
		{
			name:        "invalid format - missing colon",
			content:     `web npm start`,
			wantErr:     true,
			errContains: "invalid procfile format",
		},
		{
			name: "invalid format - empty process name",
			content: `: npm start
`,
			wantErr:     true,
			errContains: "process name cannot be empty",
		},
		{
			name: "invalid format - empty command",
			content: `web:
`,
			wantErr:     true,
			errContains: "command cannot be empty",
		},
		{
			name: "duplicate process names",
			content: `web: npm start
web: npm run dev`,
			wantErr:     true,
			errContains: "duplicate process name",
		},
		{
			name:        "empty procfile",
			content:     "",
			wantErr:     true,
			errContains: "no valid process definitions",
		},
		{
			name: "only comments",
			content: `# Just comments
# Nothing else`,
			wantErr:     true,
			errContains: "no valid process definitions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			procfilePath := filepath.Join(tmpDir, "Procfile")
			err := os.WriteFile(procfilePath, []byte(tt.content), 0644)
			assert.NilError(t, err)

			// Parse the Procfile
			procfile, err := Parse(procfilePath)

			if tt.wantErr {
				assert.ErrorContains(t, err, tt.errContains)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, len(tt.expected), len(procfile.Processes))

			for name, expectedCmd := range tt.expected {
				proc, exists := procfile.Processes[name]
				assert.Assert(t, exists, "expected process %q to exist", name)
				assert.Equal(t, expectedCmd, proc.Command)
				assert.Equal(t, name, proc.Name)
			}

			assert.Equal(t, procfilePath, procfile.FilePath)
		})
	}
}

func TestFindProcfile(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	err := os.MkdirAll(subDir, 0755)
	assert.NilError(t, err)

	// Create a Procfile in the root temp directory
	procfilePath := filepath.Join(tmpDir, "Procfile")
	err = os.WriteFile(procfilePath, []byte("web: npm start"), 0644)
	assert.NilError(t, err)

	// Test finding from subdirectory (should find parent)
	found, err := FindProcfile(subDir)
	assert.NilError(t, err)
	assert.Equal(t, procfilePath, found)

	// Test finding from the directory containing the Procfile
	found, err = FindProcfile(tmpDir)
	assert.NilError(t, err)
	assert.Equal(t, procfilePath, found)

	// Test not finding when no Procfile exists
	noProc := t.TempDir()
	_, err = FindProcfile(noProc)
	assert.ErrorContains(t, err, "Procfile not found")
}

func TestParseFromDir(t *testing.T) {
	tmpDir := t.TempDir()
	procfilePath := filepath.Join(tmpDir, "Procfile")
	content := `web: npm start
worker: node worker.js`
	err := os.WriteFile(procfilePath, []byte(content), 0644)
	assert.NilError(t, err)

	procfile, err := ParseFromDir(tmpDir)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(procfile.Processes))
	assert.Equal(t, "npm start", procfile.Processes["web"].Command)
	assert.Equal(t, "node worker.js", procfile.Processes["worker"].Command)
}
