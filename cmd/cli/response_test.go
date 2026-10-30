package cli_test

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func responseBody(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func outputMode(mode string, args []string) []string {
	if mode == "default" {
		return args
	}
	return append([]string{"--" + mode}, args...)
}

func TestDataCommandsRejectMissingOrMalformedObjects(t *testing.T) {
	commands := [][]string{
		{"user", "get"},
		{"whoami"},
		{"user", "update", "--name", "Name"},
		{"user", "ssh-keys", "add", "--key", publicKey},
		{"ssh-keys", "add", "--key", publicKey},
		{"api-keys", "get", "--prefix", "prefix"},
		{"api-keys", "create"},
		{"api-keys", "update", "--prefix", "prefix", "--label", "Label"},
		{"team", "get", "--handle", "team"},
		{"team", "create", "--handle", "team", "--name", "Name"},
		{"team", "update", "--handle", "team", "--name", "Name"},
		{"team", "accept", "--handle", "team"},
		{"team", "balance", "--handle", "team"},
		{"team", "members", "list", "--handle", "team"},
		{"team", "members", "update", "--handle", "team", "--email", "user@example.com", "--role", "user"},
		{"vm", "get", "--vm", "vm"},
		{"vm", "provision", "--gpu-model", "MI300X", "--gpu-count", "8"},
		{"vm", "state", "--vm", "vm"},
		{"vm", "wait", "--vm", "vm", "--timeout", "100ms"},
		{"bm", "get", "--server", "server"},
		{"bm", "reserve", "--cpu-cores", "128", "--ram-gb", "1024", "--disk-gb", "1000"},
		{"bm", "reinstall", "--server", "server"},
		{"bm", "console", "--server", "server"},
		{"bm", "power", "status", "--server", "server"},
		{"ssh", "--vm", "vm"},
		{"ssh", "--server", "server"},
		{"vm", "ssh", "--vm", "vm"},
		{"bm", "ssh", "--server", "server"},
		{"setup", "--non-interactive", "--team", "team"},
		{"doctor"},
	}
	for _, args := range commands {
		for _, reply := range []struct {
			status int
			body   string
		}{{200, "null"}, {200, ""}, {200, " \n "}, {204, ""}, {200, "{}"}, {200, "[]"}, {200, `{"error":"secret-response-body"}`}, {200, "<html>secret-response-body</html>"}} {
			for _, mode := range []string{"default", "json", "agent"} {
				t.Run(strings.Join(args[:min(2, len(args))], " ")+fmt.Sprintf("/%d/%q/%s", reply.status, reply.body, mode), func(t *testing.T) {
					isolatedHome(t)
					t.Setenv("HOTAISLE_API_TOKEN", "token")
					t.Setenv("HOTAISLE_DEFAULT_TEAM", "team")
					code, stdout, stderr := execute(t, t.Context(), outputMode(mode, args), func(req *http.Request) (*http.Response, error) {
						if req.URL.Path == "/api/user/ssh_keys/" && req.Method == http.MethodGet {
							return responseBody(200, "[]"), nil
						}
						return responseBody(reply.status, reply.body), nil
					})
					require.Equal(t, 1, code, stderr)
					assert.Empty(t, stdout)
					assert.Contains(t, stderr, "invalid API response")
					assert.NotContains(t, stderr, "secret-response-body")
					if mode != "default" {
						assert.Contains(t, stderr, "invalid_response")
						assert.Contains(t, stderr, "http_status")
					}
				})
			}
		}
	}
}

func TestNoContentMutationsAlwaysAcknowledgeSuccessAndReportErrors(t *testing.T) {
	commands := [][]string{
		{"vm", "update", "--vm", "vm", "--description", "Description"},
		{"vm", "delete", "--vm", "vm"},
		{"vm", "start", "--vm", "vm"},
		{"vm", "stop", "--vm", "vm"},
		{"vm", "shutdown", "--vm", "vm"},
		{"vm", "reboot", "--vm", "vm"},
		{"vm", "hard-reset", "--vm", "vm"},
		{"vm", "rebuild", "--vm", "vm"},
		{"bm", "update", "--server", "server", "--description", "Description"},
		{"bm", "delete", "--server", "server"},
		{"bm", "support-access", "enable", "--server", "server"},
		{"bm", "support-access", "disable", "--server", "server"},
		{"ssh-keys", "delete", "--fingerprint", "SHA256:key"},
		{"api-keys", "delete", "--prefix", "prefix"},
		{"team", "members", "invite", "--handle", "team", "--email", "user@example.com", "--name", "Name"},
		{"team", "members", "remove", "--handle", "team", "--email", "user@example.com"},
	}
	for _, action := range []string{"on", "shutdown", "force-shutdown", "reboot", "cold-reboot", "ac-reset"} {
		commands = append(commands, []string{"bm", "power", action, "--server", "server"})
	}
	for _, args := range commands {
		for _, status := range []int{200, 204, 403, 500} {
			for _, mode := range []string{"default", "json", "agent"} {
				t.Run(strings.Join(args[:2], " ")+fmt.Sprintf("/%d/%s", status, mode), func(t *testing.T) {
					isolatedHome(t)
					t.Setenv("HOTAISLE_DEFAULT_TEAM", "team")
					code, stdout, stderr := execute(t, t.Context(), outputMode(mode, args), func(*http.Request) (*http.Response, error) {
						return test.NewEmptyResponse(status), nil
					})
					if status >= 400 {
						assert.NotZero(t, code)
						assert.Empty(t, stdout)
						assert.Contains(t, stderr, http.StatusText(status))
						return
					}
					require.Zero(t, code, stderr)
					assert.Empty(t, stderr)
					assert.NotEmpty(t, stdout)
					assert.NotContains(t, stdout, "null")
					if mode == "json" {
						var result struct {
							Success bool   `json:"success"`
							Message string `json:"message"`
						}
						require.NoError(t, json.Unmarshal([]byte(stdout), &result))
						assert.True(t, result.Success)
						assert.NotEmpty(t, result.Message)
					}
				})
			}
		}
	}
}

func TestZeroBalancesRemainValidData(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--json", "team", "balance", "--handle", "team"}, func(*http.Request) (*http.Response, error) {
		return responseBody(200, `{"available_balance":0,"hourly_rate":0,"estimated_runout_time":null,"future_api_field":"allowed"}`), nil
	})
	require.Zero(t, code, stderr)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, `"available_balance": 0`)
	assert.Contains(t, stdout, `"hourly_rate": 0`)
}

func TestResponseValidationPreservesCaseInsensitiveFields(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--json", "vm", "get", "--team", "team", "--vm", "vm"}, func(*http.Request) (*http.Response, error) {
		return responseBody(200, `{"Name":"vm","future_api_field":true}`), nil
	})
	require.Zero(t, code, stderr)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, `"name": "vm"`)
}

func TestUnexpectedHTTPStatusAndInvalidFieldTypesAreErrors(t *testing.T) {
	for _, reply := range []struct {
		status int
		body   string
	}{{302, ""}, {200, `{"name":{"token":"secret-response-body"}}`}} {
		t.Run(fmt.Sprint(reply.status), func(t *testing.T) {
			isolatedHome(t)
			code, stdout, stderr := execute(t, t.Context(), []string{"--json", "vm", "get", "--team", "team", "--vm", "vm"}, func(*http.Request) (*http.Response, error) {
				return responseBody(reply.status, reply.body), nil
			})
			assert.Equal(t, 1, code)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "invalid_response")
			assert.NotContains(t, stderr, "secret-response-body")
		})
	}
}
