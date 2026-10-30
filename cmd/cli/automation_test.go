package cli_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hotaisle/hotaisle-cli/client"
	appcli "github.com/hotaisle/hotaisle-cli/cmd/cli"
	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("HOTAISLE_API_TOKEN", "")
	t.Setenv("HOTAISLE_CONFIG_FILE", "")
	t.Setenv("HOTAISLE_DEFAULT_TEAM", "")
	return home
}

func execute(t *testing.T, ctx context.Context, args []string, handler func(*http.Request) (*http.Response, error)) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := appcli.Execute(ctx, append([]string{"hotaisle"}, args...), &stdout, &stderr,
		client.WithHTTPClient(test.NewMockHTTPClient(handler)))
	return code, stdout.String(), stderr.String()
}

func noNetwork(t *testing.T) func(*http.Request) (*http.Response, error) {
	t.Helper()
	return func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected API request")
		return nil, nil
	}
}

func saveConfig(t *testing.T, path, token, team string) {
	t.Helper()
	data, err := json.Marshal(map[string]string{"api_token": token, "default_team": team})
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func TestHelpAndVersionDoNotReadOrWriteConfig(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--version"}, {"vm", "provision", "--help"}, {"help", "vm"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			home := isolatedHome(t)
			t.Setenv("HOTAISLE_CONFIG_FILE", filepath.Join(home, "broken.json"))
			require.NoError(t, os.WriteFile(filepath.Join(home, "broken.json"), []byte("{broken"), 0o600))
			code, stdout, _ := execute(t, t.Context(), args, noNetwork(t))
			require.Zero(t, code)
			assert.NotEmpty(t, stdout)
			_, err := os.Stat(filepath.Join(home, ".hotaisle"))
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestConfigOverridesSupplyCredentialsAndTeam(t *testing.T) {
	for _, source := range []string{"flag", "config_env", "token_env"} {
		t.Run(source, func(t *testing.T) {
			home := isolatedHome(t)
			saveConfig(t, filepath.Join(home, ".hotaisle/config.json"), "default-token", "default-team")
			custom := filepath.Join(home, "custom.json")
			saveConfig(t, custom, "custom-token", "custom-team")
			args := []string{}
			expectedToken, expectedTeam := "custom-token", "custom-team"
			switch source {
			case "flag":
				args = append(args, "--config-file", custom)
			case "config_env":
				t.Setenv("HOTAISLE_CONFIG_FILE", custom)
			default:
				t.Setenv("HOTAISLE_API_TOKEN", "  env-token  ")
				t.Setenv("HOTAISLE_DEFAULT_TEAM", "env-team")
				expectedToken, expectedTeam = "env-token", "env-team"
			}
			args = append(args, "vm", "list")
			code, _, stderr := execute(t, t.Context(), args, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, expectedToken, req.Header.Get("Authorization"))
				assert.Equal(t, "/api/teams/"+expectedTeam+"/virtual_machines/", req.URL.Path)
				return test.NewJSONResponse(t, 200, []client.VirtualMachineDetails{}), nil
			})
			require.Zero(t, code, stderr)
			assert.Empty(t, stderr)
		})
	}
}

func TestConfigWritesSelectedFileAndPreservesDefault(t *testing.T) {
	home := isolatedHome(t)
	defaultPath := filepath.Join(home, ".hotaisle/config.json")
	saveConfig(t, defaultPath, "default-token", "default-team")
	before, err := os.ReadFile(defaultPath)
	require.NoError(t, err)
	custom := filepath.Join(home, "new", "custom.json")
	code, _, stderr := execute(t, t.Context(), []string{"-c", custom, "config", "set", "default-team", "custom-team"}, noNetwork(t))
	require.Zero(t, code, stderr)
	data, err := os.ReadFile(custom)
	require.NoError(t, err)
	assert.Contains(t, string(data), "custom-team")
	after, err := os.ReadFile(defaultPath)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestEnvironmentOverridesAreNotSavedByOtherConfigCommands(t *testing.T) {
	home := isolatedHome(t)
	path := filepath.Join(home, ".hotaisle", "config.json")
	saveConfig(t, path, "stored-token", "stored-team")
	t.Setenv("HOTAISLE_API_TOKEN", "ephemeral-token")
	t.Setenv("HOTAISLE_DEFAULT_TEAM", "ephemeral-team")
	code, _, stderr := execute(t, t.Context(), []string{"config", "set", "log-level", "debug"}, noNetwork(t))
	require.Zero(t, code, stderr)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "ephemeral")
	assert.Contains(t, string(data), "stored-token")
	assert.Contains(t, string(data), "stored-team")
}

func TestExplicitTeamOverridesEnvironment(t *testing.T) {
	isolatedHome(t)
	t.Setenv("HOTAISLE_DEFAULT_TEAM", "env-team")
	code, _, stderr := execute(t, t.Context(), []string{"vm", "list", "--team", "explicit-team"}, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/api/teams/explicit-team/virtual_machines/", req.URL.Path)
		return test.NewJSONResponse(t, 200, []client.VirtualMachineDetails{}), nil
	})
	require.Zero(t, code, stderr)
}
