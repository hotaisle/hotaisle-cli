package cli_test

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
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

const publicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA test"

func TestQuickstartUsesInstalledVersionOffline(t *testing.T) {
	home := isolatedHome(t)
	path := filepath.Join(home, "broken.json")
	require.NoError(t, os.WriteFile(path, []byte("{broken"), 0o600))
	t.Setenv("HOTAISLE_CONFIG_FILE", path)
	version, commit := appcli.Version, appcli.Commit
	appcli.Version, appcli.Commit = "9.8.7", "test-commit"
	t.Cleanup(func() { appcli.Version, appcli.Commit = version, commit })
	for _, args := range [][]string{{"quickstart"}, {"--agent", "quickstart"}, {"skill"}, {"--json", "quickstart"}} {
		code, stdout, stderr := execute(t, t.Context(), args, noNetwork(t))
		require.Zero(t, code, stderr)
		assert.Contains(t, stdout, "9.8.7")
		assert.Contains(t, stdout, "test-commit")
		assert.Contains(t, stdout, "ssh-keys")
		if args[0] == "--json" {
			var result struct {
				Version  string `json:"version"`
				Markdown string `json:"markdown"`
			}
			require.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, "9.8.7", result.Version)
			assert.True(t, strings.HasPrefix(result.Markdown, "---\nname: hotaisle\n"))
		} else {
			assert.True(t, strings.HasPrefix(stdout, "---\nname: hotaisle\n"))
		}
	}
	_, err := os.Stat(filepath.Join(home, ".hotaisle"))
	assert.True(t, os.IsNotExist(err))
}

func TestProminentCommandsAppearInHelp(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--help"}, noNetwork(t))
	require.Zero(t, code, stderr)
	for _, name := range []string{"login", "ssh-keys", "api-keys", "whoami", "rentals", "offers", "quickstart", "--agent"} {
		assert.Contains(t, stdout, name)
	}
}

func TestTopLevelAccountCommandsUseExistingAPI(t *testing.T) {
	for _, tc := range []struct {
		args []string
		path string
	}{
		{[]string{"whoami"}, "/api/user/"},
		{[]string{"ssh-keys", "list"}, "/api/user/ssh_keys/"},
		{[]string{"api-keys", "list"}, "/api/user/api_keys/"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			isolatedHome(t)
			code, _, stderr := execute(t, t.Context(), tc.args, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, tc.path, req.URL.Path)
				assert.Equal(t, http.MethodGet, req.Method)
				if tc.args[0] == "whoami" {
					return test.NewJSONResponse(t, 200, client.GetUserResponse{}), nil
				}
				return test.NewJSONResponse(t, 200, []any{}), nil
			})
			require.Zero(t, code, stderr)
		})
	}
}

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
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "whoami"}, func(*http.Request) (*http.Response, error) {
		return test.NewEmptyResponse(403), nil
	})
	assert.Equal(t, 3, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "# Hot Aisle Error")
	assert.Contains(t, stderr, "authentication")
	assert.Contains(t, stderr, "403")
	for _, args := range [][]string{{"--agent", "--json", "whoami"}, {"--agent", "ssh"}, {"--agent", "--unknown"}, {"--agent", "help", "--unknown"}, {"--agent", "vm", "help", "unknown"}} {
		code, stdout, stderr := execute(t, t.Context(), args, noNetwork(t))
		assert.Equal(t, 2, code, stderr)
		assert.Empty(t, stdout)
		assert.NotContains(t, stderr, "Incorrect Usage")
	}
}

func TestAgentLoginNeverPromptsAndUsesSetup(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "login"}, noNetwork(t))
	assert.Equal(t, 3, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "no API token configured")
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

func rentalResponse(t *testing.T, req *http.Request) (*http.Response, error) {
	t.Helper()
	assert.Equal(t, http.MethodGet, req.Method)
	switch req.URL.Path {
	case "/api/teams/team/virtual_machines/available/":
		return test.NewJSONResponse(t, 200, []client.AvailableVirtualMachineTypes{{
			Quantity: 3, OnDemandPrice: 125, MinimumReservationMinutes: 30,
			Specs: client.VirtualMachineSpecs{GPUs: []client.GPUs{{Model: "MI300X", Count: 1}}, CPUCores: new(uint64(16)), RAMCapacity: new(uint64(64 << 30)), DiskCapacity: new(uint64(512 << 30))},
		}}), nil
	case "/api/teams/team/bare_metal/available/":
		return test.NewJSONResponse(t, 200, []client.AvailableBareMetalTypes{{
			Quantity: 2, OnDemandPrice: 1000, MinimumReservationMinutes: 480,
			Specs: client.BareMetalServerSpecs{GPUs: []client.GPUs{{Model: "MI300X", Count: 8}}, CPUCores: 256, RAMCapacity: 1536 << 30, DiskCapacity: 4096 << 30},
		}}), nil
	default:
		t.Fatalf("unexpected rental endpoint: %s", req.URL.Path)
		return nil, nil
	}
}

func TestOffersCombineWholeRentalBundlesAndMinimumRequirements(t *testing.T) {
	for _, args := range [][]string{{"offers", "--team", "team"}, {"--agent", "rentals", "list", "--team", "team"}, {"--json", "rentals", "--team", "team"}} {
		isolatedHome(t)
		code, stdout, stderr := execute(t, t.Context(), args, func(req *http.Request) (*http.Response, error) { return rentalResponse(t, req) })
		require.Zero(t, code, stderr)
		assert.Contains(t, stdout, "8x MI300X")
		assert.Contains(t, stdout, "1x MI300X")
		assert.Contains(t, stdout, "480")
		if args[0] == "--json" {
			var result struct {
				Currency string `json:"currency"`
				Options  []struct {
					GPURequirement   string `json:"gpu_requirement"`
					HourlyPriceCents int64  `json:"hourly_price_cents"`
				} `json:"options"`
			}
			require.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, "USD", result.Currency)
			require.Len(t, result.Options, 2)
			assert.Equal(t, int64(1000), result.Options[1].HourlyPriceCents)
		} else {
			assert.Contains(t, stdout, "1.25")
			assert.Contains(t, stdout, "10.00")
			assert.Contains(t, stdout, "1536.00")
			assert.Contains(t, stdout, "complete rental bundle")
		}
	}
}

func TestOffersFilterKindsAndDoNotPresentMissingPricesAsFree(t *testing.T) {
	isolatedHome(t)
	calls := 0
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "offers", "--team", "team", "--kind", "bm"}, func(req *http.Request) (*http.Response, error) {
		calls++
		assert.Contains(t, req.URL.Path, "bare_metal")
		return test.NewJSONResponse(t, 200, []client.AvailableBareMetalTypes{{Quantity: 1, Specs: client.BareMetalServerSpecs{GPUs: []client.GPUs{{Model: "MI300X", Count: 8}}}}}), nil
	})
	require.Zero(t, code, stderr)
	assert.Equal(t, 1, calls)
	assert.Contains(t, stdout, "not quoted")
	code, stdout, stderr = execute(t, t.Context(), []string{"--agent", "offers", "--team", "team", "--kind", "bogus"}, noNetwork(t))
	assert.Equal(t, 2, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "--kind")
}

func TestSSHKeyRegistrationReusesAccountKeys(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			home := isolatedHome(t)
			file := filepath.Join(home, "key.pub")
			require.NoError(t, os.WriteFile(file, []byte(publicKey), 0o600))
			posts := 0
			code, stdout, stderr := execute(t, t.Context(), []string{"--json", "ssh-keys", "add", file}, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "/api/user/ssh_keys/", req.URL.Path)
				key := client.SSHKey{Type: "ssh-ed25519", PublicKey: strings.Fields(publicKey)[1], Fingerprint: "SHA256:test", Comment: "different comment"}
				if req.Method == http.MethodPost {
					posts++
					var body client.SSHKeyRequest
					data, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					require.NoError(t, json.Unmarshal(data, &body))
					assert.Equal(t, publicKey, body.AuthorizedKey)
					return test.NewJSONResponse(t, 200, key), nil
				}
				keys := []client.SSHKey{}
				if existing {
					keys = append(keys, key)
				}
				return test.NewJSONResponse(t, 200, keys), nil
			})
			require.Zero(t, code, stderr)
			if existing {
				assert.Zero(t, posts)
			} else {
				assert.Equal(t, 1, posts)
			}
			assert.Contains(t, stdout, `"already_registered": `+fmt.Sprint(existing))
			assert.Contains(t, stdout, "including running instances")
			assert.Contains(t, stdout, `"guest_login_verified": false`)
		})
	}
}

func TestSSHKeyRegistrationChecksExistingInstanceBeforeMutationAndRefreshesConnection(t *testing.T) {
	isolatedHome(t)
	requests := []string{}
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "ssh-keys", "add", "--key", publicKey, "--team", "team", "--vm", "running-vm"}, func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.Method+" "+req.URL.Path)
		switch req.URL.Path {
		case "/api/user/":
			return test.NewJSONResponse(t, 200, client.GetUserResponse{Teams: []client.UserTeam{{Handle: "team", EffectiveRoles: []string{"user"}}}}), nil
		case "/api/teams/team/virtual_machines/running-vm/":
			return test.NewJSONResponse(t, 200, client.VirtualMachineDetails{Name: "running-vm", SSHAccess: &client.ExternalService{DNSName: "vm.example.com", Port: 22}}), nil
		case "/api/user/ssh_keys/":
			if req.Method == http.MethodGet {
				return test.NewJSONResponse(t, 200, []client.SSHKey{}), nil
			}
			return test.NewJSONResponse(t, 200, client.SSHKey{Fingerprint: "SHA256:test"}), nil
		default:
			t.Fatalf("unexpected request: %s", req.URL.Path)
			return nil, nil
		}
	})
	require.Zero(t, code, stderr)
	assert.Equal(t, []string{"GET /api/user/", "GET /api/teams/team/virtual_machines/running-vm/", "GET /api/user/ssh_keys/", "POST /api/user/ssh_keys/", "GET /api/teams/team/virtual_machines/running-vm/"}, requests)
	assert.Contains(t, stdout, "vm.example.com")
	assert.Contains(t, stdout, "guest_login_verified")
	assert.Contains(t, stdout, "No rebuild is required")
}

func TestSSHKeyRegistrationDoesNotGrantTeamAccess(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--json", "ssh-keys", "add", "--key", publicKey, "--team", "team", "--vm", "vm"}, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/api/user/", req.URL.Path)
		return test.NewJSONResponse(t, 200, client.GetUserResponse{Teams: []client.UserTeam{{Handle: "team", EffectiveRoles: []string{"reader"}}}}), nil
	})
	assert.Equal(t, 3, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "user role")
}

func TestSSHKeyRegistrationRejectsPrivateKeysAndAmbiguousInputs(t *testing.T) {
	isolatedHome(t)
	for _, args := range [][]string{
		{"--key", "-----BEGIN OPENSSH PRIVATE KEY----- secret"},
		{"--key", publicKey, "--file", "key.pub"},
		{"--key", publicKey, "--vm", "vm", "--server", "server"},
		{},
	} {
		code, stdout, stderr := execute(t, t.Context(), append([]string{"--json", "ssh-keys", "add"}, args...), noNetwork(t))
		assert.Equal(t, 2, code)
		assert.Empty(t, stdout)
		assert.NotContains(t, stderr, "secret")
	}
}

func TestSSHKeyStdinAndOwnerPermissionHint(t *testing.T) {
	isolatedHome(t)
	var out bytes.Buffer
	cmd := appcli.NewCommand(client.WithHTTPClient(test.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet {
			return test.NewJSONResponse(t, 200, []client.SSHKey{}), nil
		}
		return test.NewEmptyResponse(403), nil
	})))
	cmd.Reader = strings.NewReader(publicKey)
	cmd.Writer, cmd.ErrWriter = &out, &out
	err := cmd.Run(t.Context(), []string{"hotaisle", "--json", "ssh-keys", "add", "--stdin"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Empty(t, out.String())
	code, _, stderr := execute(t, t.Context(), []string{"--agent", "ssh-keys", "add", "--key", publicKey}, func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet {
			return test.NewJSONResponse(t, 200, []client.SSHKey{}), nil
		}
		return test.NewEmptyResponse(403), nil
	})
	assert.Equal(t, 3, code)
	assert.Contains(t, stderr, "personal owner permission")
}

func TestMissingInstanceStopsKeyRegistrationBeforeMutation(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--json", "ssh-keys", "add", "--key", publicKey, "--team", "team", "--vm", "missing"}, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		if req.URL.Path == "/api/user/" {
			return test.NewJSONResponse(t, 200, client.GetUserResponse{Teams: []client.UserTeam{{Handle: "team", EffectiveRoles: []string{"user"}}}}), nil
		}
		assert.Equal(t, "/api/teams/team/virtual_machines/missing/", req.URL.Path)
		return test.NewEmptyResponse(404), nil
	})
	assert.Equal(t, 4, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "not_found")
}

func TestConcurrentKeyRegistrationOnlyReusesThisAccountsKey(t *testing.T) {
	for _, owned := range []bool{true, false} {
		t.Run(fmt.Sprint(owned), func(t *testing.T) {
			isolatedHome(t)
			reads := 0
			code, stdout, stderr := execute(t, t.Context(), []string{"--json", "ssh-keys", "add", "--key", publicKey}, func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodPost {
					return test.NewEmptyResponse(409), nil
				}
				reads++
				keys := []client.SSHKey{}
				if owned && reads == 2 {
					keys = append(keys, client.SSHKey{Type: "ssh-ed25519", PublicKey: strings.Fields(publicKey)[1], Fingerprint: "SHA256:test"})
				}
				return test.NewJSONResponse(t, 200, keys), nil
			})
			assert.Equal(t, 2, reads)
			if owned {
				assert.Zero(t, code, stderr)
				assert.Contains(t, stdout, `"already_registered": true`)
			} else {
				assert.Equal(t, 5, code)
				assert.Empty(t, stdout)
			}
		})
	}
}

func TestRentalAPIErrorDoesNotEmitAPartialCatalog(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--agent", "offers", "--team", "team"}, func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "virtual_machines") {
			return rentalResponse(t, req)
		}
		return test.NewEmptyResponse(503), nil
	})
	assert.Equal(t, 6, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "temporary_failure")
}

func TestRentalJSONPreservesMissingPriceAsNull(t *testing.T) {
	isolatedHome(t)
	code, stdout, stderr := execute(t, t.Context(), []string{"--json", "offers", "--team", "team", "--kind", "vm"}, func(*http.Request) (*http.Response, error) {
		return test.NewJSONResponse(t, 200, []client.AvailableVirtualMachineTypes{{Quantity: 1}}), nil
	})
	require.Zero(t, code, stderr)
	var result struct {
		Options []struct {
			Price *int64 `json:"hourly_price_cents"`
		} `json:"options"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Options, 1)
	assert.Nil(t, result.Options[0].Price)
}
