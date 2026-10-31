package cli_test

import (
	"encoding/json/v2"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputModesAcrossAccountTeamAndResourceCommands(t *testing.T) {
	user := client.User{Name: "Test User", Email: "user@example.com"}
	team := client.UserTeam{Handle: "team", Name: "Example team", Roles: []string{"owner"}, EffectiveRoles: []string{"user"}}
	member := client.TeamMember{Name: user.Name, Email: user.Email, Roles: []string{"user"}}
	key := client.SSHKey{Type: "ssh-ed25519", PublicKey: strings.Fields(publicKey)[1], Fingerprint: "SHA256:test"}
	apiKey := client.UserAPIKey{Prefix: "prefix", Label: "Test key", UserRole: "owner"}
	vm := client.VirtualMachineDetails{Name: "vm", IPAddress: "192.0.2.1", CPUCores: new(uint64(16)), RAMCapacity: new(uint64(64 << 30)), GPUs: []client.GPUs{{Model: "MI300X", Count: 1}}, SSHAccess: &client.ExternalService{DNSName: "vm.example.com", Port: 2222}}
	server := client.BareMetalServerDetails{Name: "server", IPAddress: "192.0.2.2", CPUCores: 256, RAMCapacity: 1536 << 30, GPUs: []client.GPUs{{Model: "MI300X", Count: 8}}, SSHAccess: &client.ExternalService{DNSName: "server.example.com", Port: 2222}}
	teamDetails := client.UserTeamDetails{Handle: team.Handle, Name: team.Name, Members: []client.TeamMember{member}, VirtualMachines: []client.VirtualMachine{{Name: "vm"}}}
	cases := []struct {
		args []string
		body any
		text string
	}{
		{[]string{"whoami"}, nil, "Test User"},
		{[]string{"user", "get"}, nil, "Test User"},
		{[]string{"user", "update", "--name", "Test User"}, user, "Test User"},
		{[]string{"ssh-keys", "list"}, nil, "SHA256:test"},
		{[]string{"user", "ssh-keys", "list"}, nil, "SHA256:test"},
		{[]string{"ssh-keys", "add", "--key", publicKey}, key, "Already registered"},
		{[]string{"user", "ssh-keys", "add", "--key", publicKey}, key, "SHA256:test"},
		{[]string{"api-keys", "list"}, []client.UserAPIKey{apiKey}, "Test key"},
		{[]string{"user", "api-keys", "list"}, []client.UserAPIKey{apiKey}, "Test key"},
		{[]string{"api-keys", "get", "--prefix", "prefix"}, apiKey, "Test key"},
		{[]string{"api-keys", "create"}, client.UserAPIKeyWithToken{UserRole: "owner", Token: "new-token"}, "new-token"},
		{[]string{"api-keys", "update", "--prefix", "prefix", "--label", "Test key"}, apiKey, "Test key"},
		{[]string{"team", "list"}, []client.UserTeam{team}, "Example team"},
		{[]string{"team", "create", "--handle", "team", "--name", "Example team"}, team, "Example team"},
		{[]string{"team", "get", "--handle", "team"}, teamDetails, "Example team"},
		{[]string{"team", "update", "--handle", "team", "--name", "Example team"}, team, "Example team"},
		{[]string{"team", "invitations"}, []client.UserTeam{team}, "Example team"},
		{[]string{"team", "accept", "--handle", "team"}, team, "Example team"},
		{[]string{"team", "members", "list", "--handle", "team"}, teamDetails, "user@example.com"},
		{[]string{"team", "members", "invitations", "--handle", "team"}, []client.TeamMember{member}, "user@example.com"},
		{[]string{"team", "members", "update", "--handle", "team", "--email", "user@example.com", "--role", "user"}, member, "user@example.com"},
		{[]string{"team", "balance", "--handle", "team"}, nil, "$123.45 USD"},
		{[]string{"team", "purchase-credits", "--handle", "team", "--cents", "2500"}, client.PurchaseTeamCreditsResponse{CheckoutURL: "https://example.com/checkout", ExpiresAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}, "https://example.com/checkout"},
		{[]string{"vm", "list"}, []client.VirtualMachineDetails{vm}, "vm.example.com"},
		{[]string{"vm", "get", "--vm", "vm"}, vm, "64.00 GiB"},
		{[]string{"vm", "provision", "--gpu-model", "MI300X", "--gpu-count", "1"}, vm, "MI300X"},
		{[]string{"vm", "available"}, nil, "$1.25 USD/hour"},
		{[]string{"vm", "state", "--vm", "vm"}, client.VirtualMachineState{State: "running", Host: "vm-host"}, "running"},
		{[]string{"vm", "wait", "--vm", "vm"}, client.VirtualMachineState{State: "running", Host: "vm-host"}, "running"},
		{[]string{"bm", "list"}, []client.BareMetalServerDetails{server}, "server.example.com"},
		{[]string{"bm", "get", "--server", "server"}, server, "1536.00 GiB"},
		{[]string{"bm", "reserve", "--cpu-cores", "256", "--ram-gb", "1536", "--disk-gb", "4096"}, server, "MI300X"},
		{[]string{"bm", "available"}, nil, "$10.00 USD/hour"},
		{[]string{"bm", "reinstall", "--server", "server"}, server, "server.example.com"},
		{[]string{"bm", "power", "status", "--server", "server"}, client.BareMetalServerPowerState{State: "on"}, "on"},
		{[]string{"bm", "console", "--server", "server"}, client.BareMetalServerConsoleURL{URL: "https://example.com/console"}, "https://example.com/console"},
		{[]string{"ssh", "--vm", "vm"}, vm, "ssh -p 2222"},
		{[]string{"vm", "ssh", "--vm", "vm"}, vm, "ssh -p 2222"},
		{[]string{"ssh", "--server", "server"}, server, "ssh -p 2222"},
		{[]string{"bm", "ssh", "--server", "server"}, server, "ssh -p 2222"},
		{[]string{"ssh", "--vm", "vm", "--address"}, vm, "ssh://hotaisle@vm.example.com:2222"},
		{[]string{"offers"}, nil, "complete rental bundle"},
		{[]string{"setup", "--non-interactive", "--team", "team"}, nil, "Setup complete"},
		{[]string{"doctor"}, nil, "Authentication: OK"},
		{[]string{"config", "get", "default-team"}, nil, "team"},
		{[]string{"config", "get", "token"}, nil, "saved-token"},
		{[]string{"config", "get", "log-level"}, nil, "info"},
		{[]string{"config", "set", "default-team", "team"}, nil, "Configuration saved"},
		{[]string{"config", "set", "token"}, nil, "Configuration saved"},
		{[]string{"config", "set", "log-level", "info"}, nil, "Configuration saved"},
	}
	for _, tc := range cases {
		for _, mode := range []string{"default", "json", "agent"} {
			t.Run(strings.Join(tc.args[:min(2, len(tc.args))], " ")+"/"+mode, func(t *testing.T) {
				home := isolatedHome(t)
				saveConfig(t, filepath.Join(home, ".hotaisle/config.json"), "saved-token", "team")
				t.Setenv("HOTAISLE_API_TOKEN", "test-token")
				code, stdout, stderr := execute(t, t.Context(), outputMode(mode, tc.args), func(req *http.Request) (*http.Response, error) {
					switch {
					case req.URL.Path == "/api/user/" && req.Method == http.MethodGet:
						return test.NewJSONResponse(t, 200, client.GetUserResponse{User: user, Teams: []client.UserTeam{team}}), nil
					case req.URL.Path == "/api/user/ssh_keys/" && req.Method == http.MethodGet:
						return test.NewJSONResponse(t, 200, []client.SSHKey{key}), nil
					case strings.HasSuffix(req.URL.Path, "/balance/"):
						return test.NewJSONResponse(t, 200, client.BalanceInfo{AvailableBalance: 12345, HourlyRate: 125}), nil
					case strings.HasSuffix(req.URL.Path, "/available/"):
						return rentalResponse(t, req)
					default:
						require.NotNil(t, tc.body, "unexpected request: %s", req.URL.Path)
						return test.NewJSONResponse(t, 200, tc.body), nil
					}
				})
				require.Zero(t, code, stderr)
				assert.Empty(t, stderr)
				require.NotEmpty(t, stdout)
				if mode == "json" {
					var result any
					require.NoError(t, json.Unmarshal([]byte(stdout), &result))
					assert.True(t, strings.HasSuffix(stdout, "\n"))
				} else {
					var result any
					assert.Error(t, json.Unmarshal([]byte(stdout), &result), "JSON must require --json")
					if mode == "default" {
						assert.Contains(t, stdout, tc.text)
						assert.NotContains(t, stdout, "# Hot Aisle")
						assert.NotContains(t, stdout, "| ---")
					}
				}
			})
		}
	}
}

func TestTextOutputPreservesExactCountsAndEscapesTerminalControls(t *testing.T) {
	isolatedHome(t)
	for _, args := range [][]string{{"vm", "get", "--team", "team", "--vm", "vm"}, {"vm", "list", "--team", "team"}} {
		code, stdout, stderr := execute(t, t.Context(), args, func(*http.Request) (*http.Response, error) {
			vm := client.VirtualMachineDetails{Name: "vm\nforged-row\x1b[31m", CPUCores: new(uint64(9007199254740993))}
			if args[1] == "list" {
				return test.NewJSONResponse(t, 200, []client.VirtualMachineDetails{vm}), nil
			}
			return test.NewJSONResponse(t, 200, vm), nil
		})
		require.Zero(t, code, stderr)
		assert.Contains(t, stdout, "9007199254740993")
		assert.Contains(t, stdout, `vm\nforged-row\x1b[31m`)
		assert.NotContains(t, stdout, "\nforged-row")
		assert.NotContains(t, stdout, "\x1b")
	}
}

func TestTextBalancesUseExplicitCurrencyAndPreserveZeroAndNegativeAmounts(t *testing.T) {
	for _, tc := range []struct{ body, balance, rate string }{
		{`{"available_balance":0,"hourly_rate":0}`, "$0.00 USD", "$0.00 USD/hour"},
		{`{"available_balance":-1,"hourly_rate":125}`, "-$0.01 USD", "$1.25 USD/hour"},
		{`{"available_balance":9223372036854775807,"hourly_rate":0}`, "$92233720368547758.07 USD", "$0.00 USD/hour"},
	} {
		for _, args := range [][]string{{"team", "balance", "--handle", "team"}, {"doctor"}} {
			isolatedHome(t)
			t.Setenv("HOTAISLE_API_TOKEN", "test-token")
			t.Setenv("HOTAISLE_DEFAULT_TEAM", "team")
			code, stdout, stderr := execute(t, t.Context(), args, func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/api/user/":
					return test.NewJSONResponse(t, 200, client.GetUserResponse{Teams: []client.UserTeam{{Handle: "team"}}}), nil
				case "/api/user/ssh_keys/":
					return responseBody(200, "[]"), nil
				default:
					return responseBody(200, tc.body), nil
				}
			})
			require.Zero(t, code, stderr)
			assert.Contains(t, stdout, tc.balance)
			assert.Contains(t, stdout, tc.rate)
		}
	}
}

func TestOutputModesDoNotInventOmittedTeamLimits(t *testing.T) {
	for _, mode := range []string{"default", "json", "agent"} {
		isolatedHome(t)
		code, stdout, stderr := execute(t, t.Context(), outputMode(mode, []string{"team", "list"}), func(*http.Request) (*http.Response, error) {
			return responseBody(200, `[{"handle":"team","name":"Example team","roles":["owner"]}]`), nil
		})
		require.Zero(t, code, stderr)
		assert.Contains(t, stdout, "Example team")
		assert.NotContains(t, stdout, "maximum_virtual_machines")
		assert.NotContains(t, stdout, "Maximum virtual machines")
		assert.NotContains(t, stdout, "maximum_bare_metal_servers")
		assert.NotContains(t, stdout, "Maximum bare metal servers")
	}
}

func TestDefaultErrorsUseReadableAPIResponsesAndKeepStdoutEmpty(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"whoami"}, func(*http.Request) (*http.Response, error) {
		return responseBody(403, `{"message":"Access denied","details":{"role":"owner required"}}`), nil
	})
	assert.Equal(t, 1, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "Access denied")
	assert.Contains(t, stderr, "owner required")
	assert.NotContains(t, stderr, `{"message"`)
	code, stdout, stderr = execute(t, t.Context(), []string{"vm", "get"}, noNetwork(t))
	assert.Equal(t, 1, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "--help")
}
