//go:build watch
// +build watch

package task_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile"
)

func TestFileWatcherInterval(t *testing.T) {
	const dir = "testdata/watcher_interval"
	expectedOutput := strings.TrimSpace(`
task: Started watching for tasks: default
task: [default] echo "Hello, World!"
Hello, World!
task: [default] echo "Hello, World!"
Hello, World!
	`)

	var buff bytes.Buffer
	e := &task.Executor{
		Dir:    dir,
		Stdout: &buff,
		Stderr: &buff,
		Watch:  true,
	}

	assert.NoError(t, e.Setup())
	buff.Reset()

	err := os.MkdirAll(filepathext.SmartJoin(dir, "src"), 0755)
	assert.NoError(t, err)

	err = os.WriteFile(filepathext.SmartJoin(dir, "src/a"), []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := e.Run(ctx, taskfile.Call{Task: "default"})
				if err != nil {
					return
				}
			}
		}
	}(ctx)

	time.Sleep(10 * time.Millisecond)
	err = os.WriteFile(filepathext.SmartJoin(dir, "src/a"), []byte("test updated"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(700 * time.Millisecond)
	cancel()
	assert.Equal(t, expectedOutput, strings.TrimSpace(buff.String()))
	buff.Reset()
	err = os.RemoveAll(filepathext.SmartJoin(dir, ".task"))
	assert.NoError(t, err)
	err = os.RemoveAll(filepathext.SmartJoin(dir, "src"))
	assert.NoError(t, err)
}
