/*
   Copyright 2020 Docker Compose CLI authors

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

package formatter

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/durable_oss/durablecompose/pkg/api"
)

func TestNewLogConsumer(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, true, true, false)
	assert.Assert(t, consumer != nil)
}

func TestLogConsumerLog(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, false, false)
	consumer.Log("test-container", "test message")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "test message"))
}

func TestLogConsumerErr(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, false, false)
	consumer.Err("test-container", "error message")

	output := stderr.String()
	assert.Assert(t, strings.Contains(output, "error message"))
}

func TestLogConsumerWithPrefix(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, true, false)
	consumer.Log("service1", "message 1")
	consumer.Log("service2", "message 2")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "service1"))
	assert.Assert(t, strings.Contains(output, "message 1"))
	assert.Assert(t, strings.Contains(output, "service2"))
	assert.Assert(t, strings.Contains(output, "message 2"))
}

func TestLogConsumerWithTimestamp(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, false, true)
	consumer.Log("test-container", "timestamped message")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "timestamped message"))
	// Timestamp format includes 'T' separator
	assert.Assert(t, strings.Contains(output, "T"))
}

func TestLogConsumerWithColor(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, true, true, false)
	consumer.Log("service1", "colored message")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "colored message"))
}

func TestLogConsumerMultiline(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, false, false)
	consumer.Log("test-container", "line 1\nline 2\nline 3")

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Assert(t, len(lines) >= 3)
}

func TestLogConsumerStatus(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, true, false)
	consumer.Status("test-container", "starting")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "test-container"))
	assert.Assert(t, strings.Contains(output, "starting"))
}

func TestLogConsumerWatchLogger(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, true, true, false)
	consumer.Log(api.WatchLogger, "watch event")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "watch event"))
}

func TestLogConsumerCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, false, false)
	consumer.Log("test-container", "should not appear")

	output := stdout.String()
	assert.Equal(t, output, "")
}

func TestLogConsumerNestedContainers(t *testing.T) {
	ctx := context.Background()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	consumer := NewLogConsumer(ctx, stdout, stderr, false, true, false)
	consumer.Log("parent", "parent message")
	consumer.Log("parent child", "child message")

	output := stdout.String()
	assert.Assert(t, strings.Contains(output, "parent message"))
	assert.Assert(t, strings.Contains(output, "child message"))
}
