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

package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// MaxProcessNameLength is the maximum allowed length for a process name
	MaxProcessNameLength = 128
	// MaxCommandLength is the maximum allowed length for a command
	MaxCommandLength = 8192
	// MaxEnvironmentVariables is the maximum number of environment variables
	MaxEnvironmentVariables = 1000
	// MaxProcesses is the maximum number of processes in a manager
	MaxProcesses = 1000
	// MinStopTimeout is the minimum timeout for stopping a process
	MinStopTimeout = 100 * time.Millisecond
	// MaxStopTimeout is the maximum timeout for stopping a process
	MaxStopTimeout = 5 * time.Minute
	// DefaultStopTimeout is the default timeout for stopping a process
	DefaultStopTimeout = 10 * time.Second
)

var (
	// ErrProcessNil is returned when a nil process is encountered
	ErrProcessNil = errors.New("process cannot be nil")
	// ErrProcessNameEmpty is returned when a process name is empty
	ErrProcessNameEmpty = errors.New("process name cannot be empty")
	// ErrProcessNameTooLong is returned when a process name is too long
	ErrProcessNameTooLong = fmt.Errorf("process name cannot exceed %d characters", MaxProcessNameLength)
	// ErrCommandEmpty is returned when a command is empty
	ErrCommandEmpty = errors.New("command cannot be empty")
	// ErrCommandTooLong is returned when a command is too long
	ErrCommandTooLong = fmt.Errorf("command cannot exceed %d characters", MaxCommandLength)
	// ErrTooManyEnvironmentVariables is returned when too many environment variables are provided
	ErrTooManyEnvironmentVariables = fmt.Errorf("cannot exceed %d environment variables", MaxEnvironmentVariables)
	// ErrProcessAlreadyRunning is returned when attempting to start an already running process
	ErrProcessAlreadyRunning = errors.New("process is already running")
	// ErrProcessNotRunning is returned when attempting to stop a process that is not running
	ErrProcessNotRunning = errors.New("process is not running")
	// ErrProcessAlreadyExists is returned when adding a duplicate process
	ErrProcessAlreadyExists = errors.New("process already exists")
	// ErrProcessNotFound is returned when a process is not found
	ErrProcessNotFound = errors.New("process not found")
	// ErrManagerNil is returned when a nil manager is encountered
	ErrManagerNil = errors.New("process manager cannot be nil")
	// ErrTooManyProcesses is returned when the process manager has too many processes
	ErrTooManyProcesses = fmt.Errorf("process manager cannot exceed %d processes", MaxProcesses)
	// ErrInvalidTimeout is returned when an invalid timeout is provided
	ErrInvalidTimeout = errors.New("timeout must be between 100ms and 5 minutes")
	// ErrWorkingDirNotExist is returned when the working directory doesn't exist
	ErrWorkingDirNotExist = errors.New("working directory does not exist")
	// ErrContextCanceled is returned when the context is canceled
	ErrContextCanceled = errors.New("context canceled")
)

// ProcessState represents the state of a running process
type ProcessState string

const (
	StateStarting ProcessState = "starting"
	StateRunning  ProcessState = "running"
	StateStopped  ProcessState = "stopped"
	StateFailed   ProcessState = "failed"
	StateRestart  ProcessState = "restarting"
)

// Process represents a raw OS process (non-containerized)
type Process struct {
	Name        string
	Command     string
	WorkingDir  string
	Environment map[string]string
	State       ProcessState
	ExitCode    int
	StartedAt   time.Time
	StoppedAt   time.Time

	cmd    *exec.Cmd
	mu     sync.RWMutex
	stopCh chan struct{}
}

// ProcessManager manages multiple raw processes
type ProcessManager struct {
	processes map[string]*Process
	mu        sync.RWMutex
}

// NewProcessManager creates a new ProcessManager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

// NewProcess creates a new Process instance with validation
func NewProcess(name, command, workingDir string, env map[string]string) (*Process, error) {
	// Validate inputs
	if err := validateProcessName(name); err != nil {
		return nil, err
	}

	if err := validateCommand(command); err != nil {
		return nil, err
	}

	if len(env) > MaxEnvironmentVariables {
		return nil, ErrTooManyEnvironmentVariables
	}

	// Validate working directory if provided
	if workingDir != "" {
		cleanDir := filepath.Clean(workingDir)
		if _, err := os.Stat(cleanDir); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s", ErrWorkingDirNotExist, cleanDir)
			}
			return nil, fmt.Errorf("cannot access working directory %s: %w", cleanDir, err)
		}
		workingDir = cleanDir
	}

	// Deep copy environment to prevent external modification
	envCopy := make(map[string]string, len(env))
	for k, v := range env {
		// Validate environment variable names and values
		if err := validateEnvVar(k, v); err != nil {
			return nil, fmt.Errorf("invalid environment variable %q: %w", k, err)
		}
		envCopy[k] = v
	}

	return &Process{
		Name:        name,
		Command:     command,
		WorkingDir:  workingDir,
		Environment: envCopy,
		State:       StateStarting,
		stopCh:      make(chan struct{}),
	}, nil
}

// validateProcessName validates a process name
func validateProcessName(name string) error {
	if name == "" {
		return ErrProcessNameEmpty
	}

	if len(name) > MaxProcessNameLength {
		return ErrProcessNameTooLong
	}

	// Check for null bytes
	if strings.ContainsRune(name, '\x00') {
		return errors.New("process name contains null byte")
	}

	return nil
}

// validateCommand validates a command string
func validateCommand(command string) error {
	if command == "" {
		return ErrCommandEmpty
	}

	if len(command) > MaxCommandLength {
		return ErrCommandTooLong
	}

	// Check for null bytes
	if strings.ContainsRune(command, '\x00') {
		return errors.New("command contains null byte")
	}

	return nil
}

// validateEnvVar validates an environment variable
func validateEnvVar(key, value string) error {
	if key == "" {
		return errors.New("environment variable name cannot be empty")
	}

	// Check for null bytes
	if strings.ContainsRune(key, '\x00') || strings.ContainsRune(value, '\x00') {
		return errors.New("environment variable contains null byte")
	}

	// Environment variable names shouldn't contain '='
	if strings.Contains(key, "=") {
		return errors.New("environment variable name cannot contain '='")
	}

	const maxEnvVarLength = 32768 // 32KB
	if len(key)+len(value) > maxEnvVarLength {
		return errors.New("environment variable too large")
	}

	return nil
}

// Start starts the process
func (p *Process) Start(ctx context.Context, stdout, stderr io.Writer) error {
	if p == nil {
		return ErrProcessNil
	}

	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	// Check if context is already canceled
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		return fmt.Errorf("%w: %s", ErrProcessAlreadyRunning, p.Name)
	}

	// Validate state before starting
	if p.State == StateRunning {
		return fmt.Errorf("%w: %s", ErrProcessAlreadyRunning, p.Name)
	}

	// Use shell to execute the command for proper shell expansion
	p.cmd = exec.CommandContext(ctx, "sh", "-c", p.Command)

	if p.WorkingDir != "" {
		p.cmd.Dir = p.WorkingDir
	}

	// Set up environment (defensive copy)
	p.cmd.Env = make([]string, 0, len(os.Environ())+len(p.Environment))
	p.cmd.Env = append(p.cmd.Env, os.Environ()...)
	for k, v := range p.Environment {
		p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up output streams
	if stdout != nil {
		p.cmd.Stdout = stdout
	}
	if stderr != nil {
		p.cmd.Stderr = stderr
	}

	// Set process group ID for proper signal handling
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	p.StartedAt = time.Now()
	p.State = StateStarting

	if err := p.cmd.Start(); err != nil {
		p.State = StateFailed
		return fmt.Errorf("failed to start process %s: %w", p.Name, err)
	}

	p.State = StateRunning

	// Recreate stop channel if it was closed
	select {
	case <-p.stopCh:
		p.stopCh = make(chan struct{})
	default:
	}

	// Wait for process in background
	go p.waitForProcess()

	return nil
}

// waitForProcess waits for the process to complete (internal use only)
func (p *Process) waitForProcess() {
	if p == nil || p.cmd == nil {
		return
	}

	err := p.cmd.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.StoppedAt = time.Now()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			p.ExitCode = exitErr.ExitCode()
		} else {
			p.ExitCode = -1
		}
		p.State = StateFailed
	} else {
		p.ExitCode = 0
		p.State = StateStopped
	}

	// Safely close stop channel
	select {
	case <-p.stopCh:
		// Already closed
	default:
		close(p.stopCh)
	}
}

// Stop stops the process gracefully, with a timeout before force kill
func (p *Process) Stop(timeout time.Duration) error {
	if p == nil {
		return ErrProcessNil
	}

	// Validate timeout
	if timeout < MinStopTimeout || timeout > MaxStopTimeout {
		return fmt.Errorf("%w (got %v)", ErrInvalidTimeout, timeout)
	}

	p.mu.Lock()

	if p.cmd == nil || p.cmd.Process == nil {
		p.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrProcessNotRunning, p.Name)
	}

	// Check current state
	if p.State != StateRunning && p.State != StateStarting {
		p.mu.Unlock()
		return fmt.Errorf("%w: %s (state: %s)", ErrProcessNotRunning, p.Name, p.State)
	}

	proc := p.cmd.Process
	stopCh := p.stopCh
	p.mu.Unlock()

	// Send SIGTERM for graceful shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process may have already exited
		if !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("failed to send SIGTERM to process %s: %w", p.Name, err)
		}
		// Process already done, return success
		return nil
	}

	// Wait for process to stop or timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-stopCh:
		return nil
	case <-timer.C:
		// Force kill if timeout exceeded
		if err := proc.Kill(); err != nil {
			// Process may have exited between timeout and kill
			if !errors.Is(err, os.ErrProcessDone) {
				return fmt.Errorf("failed to kill process %s: %w", p.Name, err)
			}
		}
		<-stopCh
		return nil
	}
}

// Signal sends a signal to the process
func (p *Process) Signal(sig os.Signal) error {
	if p == nil {
		return ErrProcessNil
	}

	if sig == nil {
		return errors.New("signal cannot be nil")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("%w: %s", ErrProcessNotRunning, p.Name)
	}

	if err := p.cmd.Process.Signal(sig); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("%w: %s", ErrProcessNotRunning, p.Name)
		}
		return fmt.Errorf("failed to signal process %s: %w", p.Name, err)
	}

	return nil
}

// GetState returns the current state of the process
func (p *Process) GetState() ProcessState {
	if p == nil {
		return StateFailed
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State
}

// GetPID returns the process ID
func (p *Process) GetPID() int {
	if p == nil {
		return -1
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return -1
}

// IsRunning returns true if the process is currently running
func (p *Process) IsRunning() bool {
	if p == nil {
		return false
	}
	state := p.GetState()
	return state == StateRunning || state == StateStarting
}

// Add adds a process to the manager
func (pm *ProcessManager) Add(process *Process) error {
	if pm == nil {
		return ErrManagerNil
	}

	if process == nil {
		return ErrProcessNil
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check for maximum processes
	if len(pm.processes) >= MaxProcesses {
		return ErrTooManyProcesses
	}

	// Check for duplicate (case-insensitive)
	nameLower := strings.ToLower(process.Name)
	for existingName := range pm.processes {
		if strings.ToLower(existingName) == nameLower {
			return fmt.Errorf("%w: %s (conflicts with %s)", ErrProcessAlreadyExists, process.Name, existingName)
		}
	}

	pm.processes[process.Name] = process
	return nil
}

// Get retrieves a process by name
func (pm *ProcessManager) Get(name string) (*Process, error) {
	if pm == nil {
		return nil, ErrManagerNil
	}

	if name == "" {
		return nil, ErrProcessNameEmpty
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proc, exists := pm.processes[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProcessNotFound, name)
	}
	return proc, nil
}

// Remove removes a process from the manager
func (pm *ProcessManager) Remove(name string) error {
	if pm == nil {
		return ErrManagerNil
	}

	if name == "" {
		return ErrProcessNameEmpty
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processes[name]; !exists {
		return fmt.Errorf("%w: %s", ErrProcessNotFound, name)
	}

	delete(pm.processes, name)
	return nil
}

// List returns all processes (defensive copy)
func (pm *ProcessManager) List() []*Process {
	if pm == nil {
		return nil
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	processes := make([]*Process, 0, len(pm.processes))
	for _, proc := range pm.processes {
		processes = append(processes, proc)
	}
	return processes
}

// Count returns the number of processes
func (pm *ProcessManager) Count() int {
	if pm == nil {
		return 0
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.processes)
}

// Has checks if a process exists
func (pm *ProcessManager) Has(name string) bool {
	if pm == nil || name == "" {
		return false
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.processes[name]
	return exists
}

// StartAll starts all processes
func (pm *ProcessManager) StartAll(ctx context.Context, stdout, stderr io.Writer) error {
	if pm == nil {
		return ErrManagerNil
	}

	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	pm.mu.RLock()
	processes := make([]*Process, 0, len(pm.processes))
	for _, proc := range pm.processes {
		processes = append(processes, proc)
	}
	pm.mu.RUnlock()

	if len(processes) == 0 {
		return nil
	}

	var errs []error
	for _, proc := range processes {
		// Check context before starting each process
		select {
		case <-ctx.Done():
			errs = append(errs, fmt.Errorf("context canceled before starting %s: %w", proc.Name, ctx.Err()))
			break
		default:
		}

		if err := proc.Start(ctx, stdout, stderr); err != nil {
			errs = append(errs, fmt.Errorf("process %s: %w", proc.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to start %d/%d processes: %v", len(errs), len(processes), errs)
	}

	return nil
}

// StopAll stops all processes
func (pm *ProcessManager) StopAll(timeout time.Duration) error {
	if pm == nil {
		return ErrManagerNil
	}

	// Validate timeout
	if timeout < MinStopTimeout || timeout > MaxStopTimeout {
		return fmt.Errorf("%w (got %v)", ErrInvalidTimeout, timeout)
	}

	pm.mu.RLock()
	processes := make([]*Process, 0, len(pm.processes))
	for _, proc := range pm.processes {
		processes = append(processes, proc)
	}
	pm.mu.RUnlock()

	if len(processes) == 0 {
		return nil
	}

	var errs []error
	for _, proc := range processes {
		// Only try to stop processes that are running
		if proc.IsRunning() {
			if err := proc.Stop(timeout); err != nil {
				// Don't fail on "not running" errors
				if !errors.Is(err, ErrProcessNotRunning) {
					errs = append(errs, fmt.Errorf("process %s: %w", proc.Name, err))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop %d/%d processes: %v", len(errs), len(processes), errs)
	}

	return nil
}
