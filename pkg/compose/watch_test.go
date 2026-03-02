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

package compose

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/streams"
	"github.com/jonboulle/clockwork"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/durable_oss/durablecompose/internal/sync"
	"github.com/durable_oss/durablecompose/pkg/api"
	"github.com/durable_oss/durablecompose/pkg/mocks"
	"github.com/durable_oss/durablecompose/pkg/watch"
)

type testWatcher struct {
	events chan watch.FileEvent
	errors chan error
}

func (t testWatcher) Start() error {
	return nil
}

func (t testWatcher) Close() error {
	return nil
}

func (t testWatcher) Events() chan watch.FileEvent {
	return t.events
}

func (t testWatcher) Errors() chan error {
	return t.errors
}

type stdLogger struct{}

func (s stdLogger) Log(containerName, message string) {
	fmt.Printf("%s: %s\n", containerName, message)
}

func (s stdLogger) Err(containerName, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", containerName, message)
}

func (s stdLogger) Status(containerName, msg string) {
	fmt.Printf("%s: %s\n", containerName, msg)
}

func TestWatch_Sync(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	cli := mocks.NewMockCli(mockCtrl)
	cli.EXPECT().Err().Return(streams.NewOut(os.Stderr)).AnyTimes()
	apiClient := mocks.NewMockAPIClient(mockCtrl)
	apiClient.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(client.ContainerListResult{
		Items: []container.Summary{
			testContainer("test", "123", false),
		},
	}, nil).AnyTimes()
	// we expect the image to be pruned
	apiClient.EXPECT().ImageList(gomock.Any(), client.ImageListOptions{
		Filters: make(client.Filters).
			Add("dangling", "true").
			Add("label", api.ProjectLabel+"=myProjectName"),
	}).Return(client.ImageListResult{
		Items: []image.Summary{
			{ID: "123"},
			{ID: "456"},
		},
	}, nil).Times(1)
	apiClient.EXPECT().ImageRemove(gomock.Any(), "123", client.ImageRemoveOptions{}).Times(1)
	apiClient.EXPECT().ImageRemove(gomock.Any(), "456", client.ImageRemoveOptions{}).Times(1)
	//
	cli.EXPECT().Client().Return(apiClient).AnyTimes()

	ctx, cancelFunc := context.WithCancel(t.Context())
	t.Cleanup(cancelFunc)

	proj := types.Project{
		Name: "myProjectName",
		Services: types.Services{
			"test": {
				Name: "test",
			},
		},
	}

	watcher := testWatcher{
		events: make(chan watch.FileEvent),
		errors: make(chan error),
	}

	syncer := newFakeSyncer()
	clock := clockwork.NewFakeClock()
	go func() {
		service := composeService{
			dockerCli: cli,
			clock:     clock,
		}
		rules, err := getWatchRules(&types.DevelopConfig{
			Watch: []types.Trigger{
				{
					Path:   "/sync",
					Action: "sync",
					Target: "/work",
					Ignore: []string{"ignore"},
				},
				{
					Path:   "/rebuild",
					Action: "rebuild",
				},
			},
		}, types.ServiceConfig{Name: "test"})
		assert.NilError(t, err)

		err = service.watchEvents(ctx, &proj, api.WatchOptions{
			Build: &api.BuildOptions{},
			LogTo: stdLogger{},
			Prune: true,
		}, watcher, syncer, rules)
		assert.NilError(t, err)
	}()

	watcher.Events() <- watch.NewFileEvent("/sync/changed")
	watcher.Events() <- watch.NewFileEvent("/sync/changed/sub")
	err := clock.BlockUntilContext(ctx, 3)
	assert.NilError(t, err)
	clock.Advance(watch.QuietPeriod)
	select {
	case actual := <-syncer.synced:
		require.ElementsMatch(t, []*sync.PathMapping{
			{HostPath: "/sync/changed", ContainerPath: "/work/changed"},
			{HostPath: "/sync/changed/sub", ContainerPath: "/work/changed/sub"},
		}, actual)
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}

	watcher.Events() <- watch.NewFileEvent("/rebuild")
	watcher.Events() <- watch.NewFileEvent("/sync/changed")
	err = clock.BlockUntilContext(ctx, 4)
	assert.NilError(t, err)
	clock.Advance(watch.QuietPeriod)
	select {
	case batch := <-syncer.synced:
		t.Fatalf("received unexpected events: %v", batch)
	case <-time.After(100 * time.Millisecond):
		// expected
	}
	// TODO: there's not a great way to assert that the rebuild attempt happened
}

type fakeSyncer struct {
	synced chan []*sync.PathMapping
}

func newFakeSyncer() *fakeSyncer {
	return &fakeSyncer{
		synced: make(chan []*sync.PathMapping),
	}
}

func (f *fakeSyncer) Sync(ctx context.Context, service string, paths []*sync.PathMapping) error {
	f.synced <- paths
	return nil
}

func TestWatchEventsErrorHandling(t *testing.T) {
	t.Run("handles watcher errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mockCtrl := gomock.NewController(t)
		cli := mocks.NewMockCli(mockCtrl)
		cli.EXPECT().Err().Return(streams.NewOut(os.Stderr)).AnyTimes()

		proj := types.Project{
			Name: "test-project",
			Services: types.Services{
				"test": {Name: "test"},
			},
		}

		watcher := testWatcher{
			events: make(chan watch.FileEvent),
			errors: make(chan error),
		}

		syncer := newFakeSyncer()
		service := composeService{
			dockerCli: cli,
			clock:     clockwork.NewFakeClock(),
		}

		go func() {
			testErr := fmt.Errorf("test watcher error")
			watcher.errors <- testErr
			// Keep channel open - error is only returned when channel is closed
		}()

		// Start watching in a goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- service.watchEvents(ctx, &proj, api.WatchOptions{
				LogTo: stdLogger{},
			}, watcher, syncer, []watchRule{})
		}()

		// Wait briefly to ensure the error was logged, then verify watchEvents continues
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Verify it exits properly when context is canceled
		select {
		case err := <-errChan:
			assert.NilError(t, err)
		case <-time.After(2 * time.Second):
			t.Error("watchEvents did not exit after context cancellation")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		mockCtrl := gomock.NewController(t)
		cli := mocks.NewMockCli(mockCtrl)
		cli.EXPECT().Err().Return(streams.NewOut(os.Stderr)).AnyTimes()

		proj := types.Project{
			Name: "test-project",
			Services: types.Services{
				"test": {Name: "test"},
			},
		}

		watcher := testWatcher{
			events: make(chan watch.FileEvent),
			errors: make(chan error),
		}

		syncer := newFakeSyncer()
		service := composeService{
			dockerCli: cli,
			clock:     clockwork.NewFakeClock(),
		}

		errChan := make(chan error, 1)
		go func() {
			errChan <- service.watchEvents(ctx, &proj, api.WatchOptions{
				LogTo: stdLogger{},
			}, watcher, syncer, []watchRule{})
		}()

		cancel()

		select {
		case err := <-errChan:
			assert.NilError(t, err)
		case <-time.After(2 * time.Second):
			t.Error("expected watchEvents to exit after context cancellation")
		}
	})

	t.Run("handles large batch of events", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockCtrl := gomock.NewController(t)
		cli := mocks.NewMockCli(mockCtrl)
		cli.EXPECT().Err().Return(streams.NewOut(os.Stderr)).AnyTimes()
		apiClient := mocks.NewMockAPIClient(mockCtrl)
		cli.EXPECT().Client().Return(apiClient).AnyTimes()

		proj := types.Project{
			Name: "test-project",
			Services: types.Services{
				"test": {Name: "test"},
			},
		}

		watcher := testWatcher{
			events: make(chan watch.FileEvent, 2000),
			errors: make(chan error),
		}

		syncer := newFakeSyncer()
		clock := clockwork.NewFakeClock()
		service := composeService{
			dockerCli: cli,
			clock:     clock,
		}

		go func() {
			// Send 1500 events - more than the 1000 threshold
			for i := 0; i < 1500; i++ {
				watcher.events <- watch.NewFileEvent(fmt.Sprintf("/test/file%d", i))
			}
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		go func() {
			_ = service.watchEvents(ctx, &proj, api.WatchOptions{
				LogTo: stdLogger{},
			}, watcher, syncer, []watchRule{})
		}()

		// Wait for processing
		time.Sleep(200 * time.Millisecond)
	})
}

func TestLoadDevelopmentConfig(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		service     types.ServiceConfig
		project     *types.Project
		wantErr     bool
		errContains string
	}{
		{
			name: "no x-develop extension",
			service: types.ServiceConfig{
				Name: "test",
			},
			project: &types.Project{
				WorkingDir: tempDir,
			},
			wantErr: false,
		},
		{
			name: "valid x-develop config",
			service: types.ServiceConfig{
				Name: "test",
				Extensions: map[string]interface{}{
					"x-develop": map[string]interface{}{
						"watch": []map[string]interface{}{
							{
								"path":   "./src",
								"action": "sync",
								"target": "/app/src",
							},
						},
					},
				},
			},
			project: &types.Project{
				WorkingDir: tempDir,
			},
			wantErr: false,
		},
		{
			name: "rebuild without build section",
			service: types.ServiceConfig{
				Name: "test",
				Extensions: map[string]interface{}{
					"x-develop": map[string]interface{}{
						"watch": []map[string]interface{}{
							{
								"path":   "./src",
								"action": "rebuild",
							},
						},
					},
				},
			},
			project: &types.Project{
				WorkingDir: tempDir,
			},
			wantErr:     true,
			errContains: "doesn't have a build section",
		},
		{
			name: "sync+exec without command",
			service: types.ServiceConfig{
				Name: "test",
				Extensions: map[string]interface{}{
					"x-develop": map[string]interface{}{
						"watch": []map[string]interface{}{
							{
								"path":   "./src",
								"action": "sync+exec",
							},
						},
					},
				},
			},
			project: &types.Project{
				WorkingDir: tempDir,
			},
			wantErr:     true,
			errContains: "without a command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadDevelopmentConfig(tt.service, tt.project)
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.errContains)
			} else {
				assert.NilError(t, err)
				if tt.service.Extensions == nil {
					assert.Assert(t, config == nil)
				}
			}
		})
	}
}

func TestCheckIfPathAlreadyBindMounted(t *testing.T) {
	tests := []struct {
		name      string
		watchPath string
		volumes   []types.ServiceVolumeConfig
		want      bool
	}{
		{
			name:      "no volumes",
			watchPath: "/test/path",
			volumes:   []types.ServiceVolumeConfig{},
			want:      false,
		},
		{
			name:      "path is bind mounted",
			watchPath: "/host/src/app",
			volumes: []types.ServiceVolumeConfig{
				{
					Type:   "bind",
					Source: "/host/src",
					Target: "/app",
					Bind:   &types.ServiceVolumeBind{},
				},
			},
			want: true,
		},
		{
			name:      "path is not bind mounted",
			watchPath: "/host/other",
			volumes: []types.ServiceVolumeConfig{
				{
					Type:   "bind",
					Source: "/host/src",
					Target: "/app",
					Bind:   &types.ServiceVolumeBind{},
				},
			},
			want: false,
		},
		{
			name:      "volume is not bind type",
			watchPath: "/test/path",
			volumes: []types.ServiceVolumeConfig{
				{
					Type:   "volume",
					Source: "test-volume",
					Target: "/app",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkIfPathAlreadyBindMounted(tt.watchPath, tt.volumes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSync(t *testing.T) {
	tests := []struct {
		name    string
		trigger types.Trigger
		want    bool
	}{
		{
			name:    "sync action",
			trigger: types.Trigger{Action: types.WatchActionSync},
			want:    true,
		},
		{
			name:    "sync+restart action",
			trigger: types.Trigger{Action: types.WatchActionSyncRestart},
			want:    true,
		},
		{
			name:    "rebuild action",
			trigger: types.Trigger{Action: types.WatchActionRebuild},
			want:    false,
		},
		{
			name:    "sync+exec action",
			trigger: types.Trigger{Action: types.WatchActionSyncExec},
			want:    false,
		},
		{
			name:    "empty action",
			trigger: types.Trigger{Action: ""},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSync(tt.trigger)
			assert.Equal(t, tt.want, got)
		})
	}
}
