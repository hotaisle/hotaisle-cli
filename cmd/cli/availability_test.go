package cli_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyAvailabilityExplainsResultInEachOutputMode(t *testing.T) {
	for _, resource := range []struct{ command, endpoint, label string }{
		{"bm", "bare_metal", "bare metal servers"},
		{"vm", "virtual_machines", "virtual machines"},
	} {
		for _, body := range []string{"null", "[]", ""} {
			for _, mode := range []string{"default", "json", "agent"} {
				t.Run(resource.command+"/"+body+"/"+mode, func(t *testing.T) {
					isolatedHome(t)
					args := []string{resource.command, "available", "--team", "example-team-1"}
					if mode != "default" {
						args = append([]string{"--" + mode}, args...)
					}
					code, stdout, stderr := execute(t, t.Context(), args, func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodGet, req.Method)
						assert.Equal(t, "/api/teams/example-team-1/"+resource.endpoint+"/available/", req.URL.Path)
						return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					require.Zero(t, code, stderr)
					assert.Empty(t, stderr)
					switch mode {
					case "json":
						assert.Equal(t, "[]\n", stdout)
					case "agent":
						assert.Contains(t, stdout, "# Hot Aisle Availability")
						assert.Contains(t, stdout, "No "+resource.label+" are currently available")
						assert.Contains(t, stdout, "example-team-1")
						assert.Contains(t, stdout, "count")
					default:
						assert.Equal(t, "No "+resource.label+" are currently available for team \"example-team-1\".\n", stdout)
					}
				})
			}
		}
	}
}

func TestAvailabilityDoesNotHidePermissionErrorsAsEmptyResults(t *testing.T) {
	for _, resource := range []string{"bm", "vm"} {
		t.Run(resource, func(t *testing.T) {
			isolatedHome(t)
			code, stdout, stderr := execute(t, t.Context(), []string{"--json", resource, "available", "--team", "team"}, func(*http.Request) (*http.Response, error) {
				return test.NewEmptyResponse(http.StatusForbidden), nil
			})
			assert.Equal(t, 3, code)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "authentication")
			assert.NotContains(t, stderr, "currently available")
		})
	}
}
