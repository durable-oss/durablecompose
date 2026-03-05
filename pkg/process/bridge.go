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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/durable_oss/durablecompose/pkg/config"
	"github.com/durable_oss/durablecompose/pkg/procfile"
)

// LoadProcessesFromConfig loads process definitions from DurableComposeConfig
func LoadProcessesFromConfig(cfg *config.DurableComposeConfig, workDir string) (*ProcessManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if !cfg.Processes.Enabled {
		return NewProcessManager(), nil
	}

	pm := NewProcessManager()

	// Determine working directory
	processWorkDir := workDir
	if cfg.Processes.WorkingDir != "" {
		processWorkDir = cfg.Processes.WorkingDir
		if !filepath.IsAbs(processWorkDir) {
			processWorkDir = filepath.Join(workDir, processWorkDir)
		}
	}

	// Load from Procfile if specified
	if cfg.Processes.ProcfilePath != "" {
		procfilePath := cfg.Processes.ProcfilePath
		if !filepath.IsAbs(procfilePath) {
			procfilePath = filepath.Join(workDir, procfilePath)
		}

		pf, err := procfile.Parse(procfilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Procfile: %w", err)
		}

		for _, proc := range pf.Processes {
			env := make(map[string]string)
			// Copy default environment
			for k, v := range cfg.Processes.Environment {
				env[k] = v
			}

			process, err := NewProcess(proc.Name, proc.Command, processWorkDir, env)
			if err != nil {
				return nil, err
			}
			if err := pm.Add(process); err != nil {
				return nil, err
			}
		}
	}

	// Load from inline definitions (can override Procfile)
	for name, def := range cfg.Processes.Definitions {
		env := make(map[string]string)
		// Copy default environment
		for k, v := range cfg.Processes.Environment {
			env[k] = v
		}
		// Apply process-specific environment
		for k, v := range def.Environment {
			env[k] = v
		}

		wd := processWorkDir
		if def.WorkingDir != "" {
			wd = def.WorkingDir
			if !filepath.IsAbs(wd) {
				wd = filepath.Join(workDir, wd)
			}
		}

		process, err := NewProcess(name, def.Command, wd, env)
		if err != nil {
			return nil, err
		}

		// Remove existing if present (allow override)
		pm.Remove(name)

		if err := pm.Add(process); err != nil {
			return nil, err
		}
	}

	return pm, nil
}

// LoadProcessesFromProcfile loads process definitions from a Procfile
func LoadProcessesFromProcfile(procfilePath string, workDir string, env map[string]string) (*ProcessManager, error) {
	pf, err := procfile.Parse(procfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Procfile: %w", err)
	}

	pm := NewProcessManager()

	for _, proc := range pf.Processes {
		// Copy environment
		procEnv := make(map[string]string)
		for k, v := range env {
			procEnv[k] = v
		}

		process, err := NewProcess(proc.Name, proc.Command, workDir, procEnv)
		if err != nil {
			return nil, err
		}
		if err := pm.Add(process); err != nil {
			return nil, err
		}
	}

	return pm, nil
}

// ProcessLogWriter wraps a writer to add process name prefix
type ProcessLogWriter struct {
	ProcessName string
	Writer      io.Writer
	IsStderr    bool
}

// Write implements io.Writer with prefix
func (w *ProcessLogWriter) Write(p []byte) (n int, err error) {
	prefix := fmt.Sprintf("[%s] ", w.ProcessName)
	if w.IsStderr {
		prefix = fmt.Sprintf("[%s] [ERROR] ", w.ProcessName)
	}

	// Write prefix and content
	_, err = w.Writer.Write([]byte(prefix))
	if err != nil {
		return 0, err
	}

	return w.Writer.Write(p)
}

// StartProcessesWithLogging starts all processes with logging to stdout/stderr
func StartProcessesWithLogging(ctx context.Context, pm *ProcessManager) error {
	processes := pm.List()

	for _, proc := range processes {
		stdout := &ProcessLogWriter{
			ProcessName: proc.Name,
			Writer:      os.Stdout,
			IsStderr:    false,
		}
		stderr := &ProcessLogWriter{
			ProcessName: proc.Name,
			Writer:      os.Stderr,
			IsStderr:    true,
		}

		if err := proc.Start(ctx, stdout, stderr); err != nil {
			return fmt.Errorf("failed to start process %s: %w", proc.Name, err)
		}
	}

	return nil
}
