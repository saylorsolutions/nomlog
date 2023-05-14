package runtime

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/plugin/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnstructured(t *testing.T) {
	r := NewRuntime(hclog.Default(), file.Plugin())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, r.Start(ctx))

	dir, err := os.MkdirTemp("", "TestUnstructured-*")
	require.NoError(t, err)
	t.Log("Using dir:", dir)
	defer func() {
		t.Log("Removing temp dir")
		err := os.RemoveAll(dir)
		if err != nil {
			t.Error("Failed to remove dir:", err)
		}
		t.Log("Done")
	}()
	defer func() {
		_ = r.Stop()
	}()
	output := filepath.Join(dir, "output.json")
	err = r.ExecuteString(`
source as src file.File "data.txt"
tag src with "unstructured"
fanout src as a and b
merge a and b as src2
dupe src2 as c and d
merge d and c as src3
sink src3 to file.File "` + output + `"
`)
	assert.NoError(t, err)

	data, err := os.ReadFile(output)
	assert.NoError(t, err)
	assert.True(t, len(data) > 0, "Data length should be greater than 0")
	t.Log(string(data))
}

func TestStructured(t *testing.T) {
	r := NewRuntime(hclog.Default(), file.Plugin())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, r.Start(ctx))

	dir, err := os.MkdirTemp("", "TestStructured-*")
	require.NoError(t, err)
	t.Log("Using dir:", dir)
	defer func() {
		t.Log("Removing temp dir")
		err := os.RemoveAll(dir)
		if err != nil {
			t.Error("Failed to remove dir:", err)
		}
		t.Log("Done")
	}()
	defer func() {
		_ = r.Stop()
	}()
	output := filepath.Join(dir, "output.json")
	err = r.ExecuteString(`
source as src file.File "data.json"
tag src with "structured"
sink src to file.File "` + output + `"
`)
	assert.NoError(t, err)

	data, err := os.ReadFile(output)
	assert.NoError(t, err)
	assert.True(t, len(data) > 0, "Data length should be greater than 0")
	t.Log(string(data))
}
