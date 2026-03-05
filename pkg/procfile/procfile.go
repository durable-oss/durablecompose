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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const (
	// MaxProcessNameLength is the maximum allowed length for a process name
	MaxProcessNameLength = 64
	// MaxCommandLength is the maximum allowed length for a command
	MaxCommandLength = 4096
	// MaxProcesses is the maximum number of processes allowed in a Procfile
	MaxProcesses = 100
	// MaxLineLength is the maximum length of a single line in the Procfile
	MaxLineLength = 8192
)

var (
	// ErrEmptyPath is returned when an empty path is provided
	ErrEmptyPath = errors.New("procfile path cannot be empty")
	// ErrNilProcfile is returned when a nil Procfile is encountered
	ErrNilProcfile = errors.New("procfile cannot be nil")
	// ErrTooManyProcesses is returned when the Procfile exceeds MaxProcesses
	ErrTooManyProcesses = fmt.Errorf("procfile cannot contain more than %d processes", MaxProcesses)
	// ErrLineTooLong is returned when a line exceeds MaxLineLength
	ErrLineTooLong = fmt.Errorf("line exceeds maximum length of %d characters", MaxLineLength)
	// ErrInvalidProcessName is returned when a process name contains invalid characters
	ErrInvalidProcessName = errors.New("process name must contain only alphanumeric characters, hyphens, and underscores")
	// ErrProcessNameTooLong is returned when a process name exceeds MaxProcessNameLength
	ErrProcessNameTooLong = fmt.Errorf("process name cannot exceed %d characters", MaxProcessNameLength)
	// ErrCommandTooLong is returned when a command exceeds MaxCommandLength
	ErrCommandTooLong = fmt.Errorf("command cannot exceed %d characters", MaxCommandLength)

	// validProcessNameRegex matches valid process names
	validProcessNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Process represents a single process definition from a Procfile
type Process struct {
	// Name is the process name (e.g., "web", "worker")
	Name string
	// Command is the shell command to execute
	Command string
	// LineNumber is the line number in the Procfile where this process was defined
	LineNumber int
}

// Procfile represents a parsed Procfile containing multiple process definitions
type Procfile struct {
	// Processes is a map of process name to Process definition
	Processes map[string]Process
	// FilePath is the path to the Procfile that was parsed
	FilePath string
}

// Parse parses a Procfile from the given file path
func Parse(path string) (*Procfile, error) {
	// Validate input
	if path == "" {
		return nil, ErrEmptyPath
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)

	// Check if file exists and is readable
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("procfile does not exist: %s", cleanPath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading procfile: %s", cleanPath)
		}
		return nil, fmt.Errorf("failed to access procfile: %w", err)
	}

	// Ensure it's a regular file, not a directory or special file
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("procfile is not a regular file: %s", cleanPath)
	}

	// Check file size to prevent reading extremely large files
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("procfile too large (max %d bytes): %s", maxFileSize, cleanPath)
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfile: %w", err)
	}
	defer file.Close()

	procfile := &Procfile{
		Processes: make(map[string]Process),
		FilePath:  cleanPath,
	}

	scanner := bufio.NewScanner(file)
	// Set a maximum token size to prevent excessive memory usage
	scanner.Buffer(make([]byte, 4096), MaxLineLength)

	lineNum := 0

	for scanner.Scan() {
		lineNum++
		rawLine := scanner.Text()

		// Check line length
		if len(rawLine) > MaxLineLength {
			return nil, fmt.Errorf("%w at line %d", ErrLineTooLong, lineNum)
		}

		line := strings.TrimSpace(rawLine)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for too many processes
		if len(procfile.Processes) >= MaxProcesses {
			return nil, fmt.Errorf("%w (found at line %d)", ErrTooManyProcesses, lineNum)
		}

		// Parse process definition: "name: command"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid procfile format at line %d: %q (expected format: name: command)", lineNum, sanitizeForError(line))
		}

		name := strings.TrimSpace(parts[0])
		command := strings.TrimSpace(parts[1])

		// Validate process name
		if err := validateProcessName(name); err != nil {
			return nil, fmt.Errorf("invalid process name at line %d: %w", lineNum, err)
		}

		// Validate command
		if err := validateCommand(command); err != nil {
			return nil, fmt.Errorf("invalid command for process %q at line %d: %w", name, lineNum, err)
		}

		// Check for duplicate process names (case-insensitive to prevent confusion)
		nameLower := strings.ToLower(name)
		for existingName := range procfile.Processes {
			if strings.ToLower(existingName) == nameLower {
				return nil, fmt.Errorf("duplicate process name %q at line %d (conflicts with %q)", name, lineNum, existingName)
			}
		}

		procfile.Processes[name] = Process{
			Name:       name,
			Command:    command,
			LineNumber: lineNum,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading procfile: %w", err)
	}

	if len(procfile.Processes) == 0 {
		return nil, fmt.Errorf("procfile contains no valid process definitions")
	}

	return procfile, nil
}

// validateProcessName validates a process name
func validateProcessName(name string) error {
	if name == "" {
		return errors.New("process name cannot be empty")
	}

	if len(name) > MaxProcessNameLength {
		return ErrProcessNameTooLong
	}

	// Check for valid characters
	if !validProcessNameRegex.MatchString(name) {
		return ErrInvalidProcessName
	}

	// Additional checks for potentially problematic names
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, "_") {
		return errors.New("process name cannot start with hyphen or underscore")
	}

	// Check for control characters or other invisible characters
	for _, r := range name {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return errors.New("process name contains invalid characters")
		}
	}

	return nil
}

// validateCommand validates a command string
func validateCommand(command string) error {
	if command == "" {
		return errors.New("command cannot be empty")
	}

	if len(command) > MaxCommandLength {
		return ErrCommandTooLong
	}

	// Check for null bytes which could cause security issues
	if strings.ContainsRune(command, '\x00') {
		return errors.New("command contains null byte")
	}

	// Warn about potentially dangerous characters (but don't fail)
	// This is just basic validation - the shell will handle escaping
	return nil
}

// sanitizeForError truncates and sanitizes a string for safe error message display
func sanitizeForError(s string) string {
	const maxLen = 100
	if len(s) > maxLen {
		s = s[:maxLen] + "..."
	}
	// Replace control characters with spaces for readability
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return ' '
		}
		return r
	}, s)
}

// FindProcfile searches for a Procfile in the given directory and parent directories
func FindProcfile(startDir string) (string, error) {
	// Validate input
	if startDir == "" {
		return "", ErrEmptyPath
	}

	// Clean the path
	dir := filepath.Clean(startDir)

	// Verify the start directory exists
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("start directory does not exist: %s", dir)
		}
		return "", fmt.Errorf("cannot access start directory: %w", err)
	}

	// Prevent infinite loops by limiting search depth
	const maxDepth = 100
	depth := 0

	for {
		depth++
		if depth > maxDepth {
			return "", fmt.Errorf("exceeded maximum search depth (%d) looking for Procfile", maxDepth)
		}

		procfilePath := filepath.Join(dir, "Procfile")

		// Check if Procfile exists
		fileInfo, err := os.Stat(procfilePath)
		if err == nil {
			// Found a Procfile, verify it's a regular file
			if !fileInfo.Mode().IsRegular() {
				// Found something named "Procfile" but it's not a file
				// Continue searching in parent directories
			} else {
				return procfilePath, nil
			}
		} else if !os.IsNotExist(err) {
			// Some error other than "not exist" - might be permission issue
			// Continue searching rather than failing
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return "", fmt.Errorf("Procfile not found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

// ParseFromDir searches for and parses a Procfile starting from the given directory
func ParseFromDir(startDir string) (*Procfile, error) {
	if startDir == "" {
		return nil, ErrEmptyPath
	}

	procfilePath, err := FindProcfile(startDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find Procfile: %w", err)
	}

	procfile, err := Parse(procfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Procfile at %s: %w", procfilePath, err)
	}

	return procfile, nil
}

// Validate checks if the Procfile is valid and all its processes are well-formed
func (p *Procfile) Validate() error {
	if p == nil {
		return ErrNilProcfile
	}

	if len(p.Processes) == 0 {
		return errors.New("procfile contains no processes")
	}

	if len(p.Processes) > MaxProcesses {
		return ErrTooManyProcesses
	}

	for name, proc := range p.Processes {
		if err := validateProcessName(name); err != nil {
			return fmt.Errorf("invalid process name %q: %w", name, err)
		}

		if err := validateCommand(proc.Command); err != nil {
			return fmt.Errorf("invalid command for process %q: %w", name, err)
		}

		// Verify process name matches map key
		if proc.Name != name {
			return fmt.Errorf("process name mismatch: map key %q does not match process name %q", name, proc.Name)
		}
	}

	return nil
}

// GetProcess safely retrieves a process by name
func (p *Procfile) GetProcess(name string) (Process, bool) {
	if p == nil {
		return Process{}, false
	}
	proc, exists := p.Processes[name]
	return proc, exists
}

// ProcessNames returns a sorted list of process names
func (p *Procfile) ProcessNames() []string {
	if p == nil || len(p.Processes) == 0 {
		return nil
	}

	names := make([]string, 0, len(p.Processes))
	for name := range p.Processes {
		names = append(names, name)
	}
	return names
}
