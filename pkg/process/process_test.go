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
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestProcess_Start(t *testing.T) {
	ctx := context.Background()

	t.Run("simple command", func(t *testing.T) {
		var stdout bytes.Buffer
		proc, err := NewProcess("test", "echo hello", "", nil)
		assert.NilError(t, err)

		err = proc.Start(ctx, &stdout, nil)
		assert.NilError(t, err)

		// Wait for process to complete
		time.Sleep(200 * time.Millisecond)

		// Process may have already finished (quick commands like echo)
		state := proc.GetState()
		assert.Assert(t, state == StateRunning || state == StateStopped || state == StateFailed)
		assert.Assert(t, strings.Contains(stdout.String(), "hello"))
	})

	t.Run("with environment variables", func(t *testing.T) {
		var stdout bytes.Buffer
		env := map[string]string{
			"TEST_VAR": "test_value",
		}
		proc, err := NewProcess("test", "echo $TEST_VAR", "", env)
		assert.NilError(t, err)

		err = proc.Start(ctx, &stdout, nil)
		assert.NilError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.Assert(t, strings.Contains(stdout.String(), "test_value"))
	})

	t.Run("with working directory", func(t *testing.T) {
		var stdout bytes.Buffer
		proc, err := NewProcess("test", "pwd", "/tmp", nil)
		assert.NilError(t, err)

		err = proc.Start(ctx, &stdout, nil)
		assert.NilError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.Assert(t, strings.Contains(stdout.String(), "/tmp"))
	})

	t.Run("invalid command", func(t *testing.T) {
		proc, err := NewProcess("test", "nonexistent_command_xyz", "", nil)
		assert.NilError(t, err)

		_ = proc.Start(ctx, nil, nil)
		// Command will start but fail - wait a bit for it to fail
		time.Sleep(100 * time.Millisecond)

		// The process should have failed
		state := proc.GetState()
		assert.Assert(t, state == StateFailed || state == StateStopped)
	})
}

func TestProcess_Stop(t *testing.T) {
	ctx := context.Background()

	t.Run("graceful stop", func(t *testing.T) {
		proc, err := NewProcess("test", "sleep 10", "", nil)
		assert.NilError(t, err)

		err = proc.Start(ctx, nil, nil)
		assert.NilError(t, err)

		// Give it time to start
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, StateRunning, proc.GetState())

		// Stop the process
		err = proc.Stop(1 * time.Second)
		assert.NilError(t, err)

		// Process should be stopped
		state := proc.GetState()
		assert.Assert(t, state == StateStopped || state == StateFailed)
	})

	t.Run("force kill on timeout", func(t *testing.T) {
		// Use a command that ignores SIGTERM
		proc, err := NewProcess("test", "trap '' TERM; sleep 10", "", nil)
		assert.NilError(t, err)

		err = proc.Start(ctx, nil, nil)
		assert.NilError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Stop with very short timeout to force kill
		start := time.Now()
		err = proc.Stop(100 * time.Millisecond)
		duration := time.Since(start)

		assert.NilError(t, err)
		// Should have been killed quickly (within a reasonable time)
		assert.Assert(t, duration < 2*time.Second)
	})
}

func TestProcess_GetPID(t *testing.T) {
	ctx := context.Background()
	proc, err := NewProcess("test", "sleep 1", "", nil)
	assert.NilError(t, err)

	// PID should be -1 before start
	assert.Equal(t, -1, proc.GetPID())

	err = proc.Start(ctx, nil, nil)
	assert.NilError(t, err)

	// PID should be valid after start
	pid := proc.GetPID()
	assert.Assert(t, pid > 0)

	// Clean up
	proc.Stop(1 * time.Second)
}

func TestProcessManager(t *testing.T) {
	pm := NewProcessManager()

	t.Run("add and get", func(t *testing.T) {
		proc, err := NewProcess("test1", "echo hello", "", nil)
		assert.NilError(t, err)
		err = pm.Add(proc)
		assert.NilError(t, err)

		retrieved, err := pm.Get("test1")
		assert.NilError(t, err)
		assert.Equal(t, proc.Name, retrieved.Name)
	})

	t.Run("duplicate add", func(t *testing.T) {
		proc, err := NewProcess("test2", "echo hello", "", nil)
		assert.NilError(t, err)
		err = pm.Add(proc)
		assert.NilError(t, err)

		// Try to add again
		err = pm.Add(proc)
		assert.ErrorContains(t, err, "already exists")
	})

	t.Run("get nonexistent", func(t *testing.T) {
		_, err := pm.Get("nonexistent")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("remove", func(t *testing.T) {
		proc, err := NewProcess("test3", "echo hello", "", nil)
		assert.NilError(t, err)
		err = pm.Add(proc)
		assert.NilError(t, err)

		err = pm.Remove("test3")
		assert.NilError(t, err)

		_, err = pm.Get("test3")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("list", func(t *testing.T) {
		pm2 := NewProcessManager()
		proc1, err := NewProcess("proc1", "echo 1", "", nil)
		assert.NilError(t, err)
		proc2, err := NewProcess("proc2", "echo 2", "", nil)
		assert.NilError(t, err)

		pm2.Add(proc1)
		pm2.Add(proc2)

		list := pm2.List()
		assert.Equal(t, 2, len(list))
	})
}

func TestProcessManager_StartStopAll(t *testing.T) {
	ctx := context.Background()
	pm := NewProcessManager()

	proc1, err := NewProcess("proc1", "sleep 10", "", nil)
	assert.NilError(t, err)
	proc2, err := NewProcess("proc2", "sleep 10", "", nil)
	assert.NilError(t, err)

	pm.Add(proc1)
	pm.Add(proc2)

	// Start all
	err = pm.StartAll(ctx, nil, nil)
	assert.NilError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Check both are running
	assert.Equal(t, StateRunning, proc1.GetState())
	assert.Equal(t, StateRunning, proc2.GetState())

	// Stop all
	err = pm.StopAll(1 * time.Second)
	assert.NilError(t, err)

	// Check both are stopped
	state1 := proc1.GetState()
	state2 := proc2.GetState()
	assert.Assert(t, state1 == StateStopped || state1 == StateFailed)
	assert.Assert(t, state2 == StateStopped || state2 == StateFailed)
}
