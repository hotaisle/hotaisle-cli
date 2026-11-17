package client_test

import (
	"net/http"
	"testing"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/hotaisle/hotaisle-cli/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthorization(t *testing.T) {
	for _, input := range []string{
		"lookup.secret", "ha_ak_lookup.secret",
		"Bearer lookup.secret", "bearer ha_ak_lookup.secret",
		"Token lookup.secret", "tOkEn ha_ak_lookup.secret",
		"  BEARER\tha_ak_lookup.secret  ", "", " \t ",
	} {
		t.Run(input, func(t *testing.T) {
			want := "Bearer ha_ak_lookup.secret"
			switch input {
			case "lookup.secret", "Bearer lookup.secret", "Token lookup.secret":
				want = "Bearer lookup.secret"
			case "", " \t ":
				want = ""
			}
			for _, setter := range []string{"WithToken", "SetToken"} {
				t.Run(setter, func(t *testing.T) {
					api := client.NewClient(client.WithHTTPClient(test.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, want, req.Header.Get("Authorization"))
						if want == "" {
							assert.Empty(t, req.Header.Values("Authorization"))
						} else {
							assert.Equal(t, []string{want}, req.Header.Values("Authorization"))
						}
						return test.NewJSONResponse(t, http.StatusOK, []client.UserAPIKey{}), nil
					})), client.WithToken("previous.secret"))
					if setter == "WithToken" {
						client.WithToken(input)(api)
					} else {
						api.SetToken(input)
					}
					_, err := api.User().GetAPIKeys(t.Context())
					require.NoError(t, err)
				})
			}
		})
	}
}
