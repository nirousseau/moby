package client // import "github.com/docker/docker/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/checkpoint"
	"github.com/docker/docker/errdefs"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestCheckpointListError(t *testing.T) {
	client := &Client{
		client: newMockClient(errorMock(http.StatusInternalServerError, "Server error")),
	}

	_, err := client.CheckpointList(context.Background(), "container_id", checkpoint.ListOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsSystem))
}

func TestCheckpointList(t *testing.T) {
	expectedURL := "/containers/container_id/checkpoints"

	client := &Client{
		client: newMockClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasPrefix(req.URL.Path, expectedURL) {
				return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
			}
			content, err := json.Marshal([]checkpoint.Summary{
				{
					Name: "checkpoint",
				},
			})
			if err != nil {
				return nil, err
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(content)),
			}, nil
		}),
	}

	checkpoints, err := client.CheckpointList(context.Background(), "container_id", checkpoint.ListOptions{})
	assert.NilError(t, err)
	assert.Check(t, is.Len(checkpoints, 1))
}

func TestCheckpointListContainerNotFound(t *testing.T) {
	client := &Client{
		client: newMockClient(errorMock(http.StatusNotFound, "Server error")),
	}

	_, err := client.CheckpointList(context.Background(), "unknown", checkpoint.ListOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsNotFound))
}
