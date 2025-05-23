package client // import "github.com/docker/docker/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/errdefs"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestServiceInspectError(t *testing.T) {
	client := &Client{
		client: newMockClient(errorMock(http.StatusInternalServerError, "Server error")),
	}

	_, _, err := client.ServiceInspectWithRaw(context.Background(), "nothing", types.ServiceInspectOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsSystem))
}

func TestServiceInspectServiceNotFound(t *testing.T) {
	client := &Client{
		client: newMockClient(errorMock(http.StatusNotFound, "Server error")),
	}

	_, _, err := client.ServiceInspectWithRaw(context.Background(), "unknown", types.ServiceInspectOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsNotFound))
}

func TestServiceInspectWithEmptyID(t *testing.T) {
	client := &Client{
		client: newMockClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("should not make request")
		}),
	}
	_, _, err := client.ServiceInspectWithRaw(context.Background(), "", types.ServiceInspectOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsInvalidParameter))
	assert.Check(t, is.ErrorContains(err, "value is empty"))

	_, _, err = client.ServiceInspectWithRaw(context.Background(), "    ", types.ServiceInspectOptions{})
	assert.Check(t, is.ErrorType(err, errdefs.IsInvalidParameter))
	assert.Check(t, is.ErrorContains(err, "value is empty"))
}

func TestServiceInspect(t *testing.T) {
	expectedURL := "/services/service_id"
	client := &Client{
		client: newMockClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasPrefix(req.URL.Path, expectedURL) {
				return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
			}
			content, err := json.Marshal(swarm.Service{
				ID: "service_id",
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

	serviceInspect, _, err := client.ServiceInspectWithRaw(context.Background(), "service_id", types.ServiceInspectOptions{})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(serviceInspect.ID, "service_id"))
}
