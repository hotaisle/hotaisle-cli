package cli_test

import (
	"net/http"
	"testing"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentOutputPreservesNumbersAndEscapesMarkdown(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "vm", "list", "--team", "team"}, func(*http.Request) (*http.Response, error) {
		return test.NewJSONResponse(t, 200, []client.VirtualMachineDetails{{Name: "quote | `\n# injected", CPUCores: new(uint64(9007199254740993))}}), nil
	})
	require.Zero(t, code, stderr)
	assert.Contains(t, stdout, "# Hot Aisle Result")
	assert.Contains(t, stdout, "9007199254740993")
	assert.Contains(t, stdout, "\\|")
	assert.NotContains(t, stdout, "\n# injected")
	assert.Empty(t, stderr)
}

func TestAgentErrorsUseMarkdownAndCategorizedExitCodes(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "user", "get"}, func(*http.Request) (*http.Response, error) {
		return test.NewEmptyResponse(403), nil
	})
	assert.Equal(t, 3, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "# Hot Aisle Error")
	assert.Contains(t, stderr, "authentication")
	assert.Contains(t, stderr, "403")
	for _, args := range [][]string{{"--agent", "--json", "user", "get"}, {"--agent", "ssh"}, {"--agent", "--unknown"}, {"--agent", "help", "--unknown"}, {"--agent", "vm", "help", "unknown"}} {
		code, stdout, stderr := execute(t, t.Context(), args, noNetwork(t))
		assert.Equal(t, 2, code, stderr)
		assert.Empty(t, stdout)
		assert.NotContains(t, stderr, "Incorrect Usage")
	}
}

func TestDirectSSHAndAddressForBothResourceKinds(t *testing.T) {
	for _, flag := range []string{"--vm", "--server"} {
		for _, mode := range []string{"command", "address", "agent"} {
			t.Run(flag+"_"+mode, func(t *testing.T) {
				isolatedHome(t)
				args := []string{"ssh", "--team", "team", flag, "instance"}
				if mode == "address" {
					args = append(args, "--address")
				}
				if mode == "agent" {
					args = append([]string{"--agent"}, args...)
				}
				code, stdout, stderr := execute(t, t.Context(), args, func(req *http.Request) (*http.Response, error) {
					kind := "virtual_machines"
					if flag == "--server" {
						kind = "bare_metal"
					}
					assert.Contains(t, req.URL.Path, "/teams/team/"+kind+"/")
					return test.NewJSONResponse(t, 200, map[string]any{"ssh_access": client.ExternalService{IPAddress: "2001:db8::1", Port: 2222}}), nil
				})
				require.Zero(t, code, stderr)
				switch mode {
				case "address":
					assert.Equal(t, "ssh://hotaisle@[2001:db8::1]:2222\n", stdout)
				case "command":
					assert.Equal(t, "ssh -p 2222 -- hotaisle@2001:db8::1\n", stdout)
				default:
					assert.Contains(t, stdout, "argv")
					assert.Contains(t, stdout, "2222")
				}
			})
		}
	}
}
