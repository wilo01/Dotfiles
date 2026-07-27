package token

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var (
	flagURL      string
	flagResource string
	flagClientID string
)

var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Generate a TDS OAuth bearer token",
	Long: `Fetch a bearer token from the TDS OAuth endpoint and print the raw JSON
response to stdout, so it pipes cleanly to jq.

The client secret is read from the HLP_TDS_CLIENT_SECRET environment variable
(falling back to TDS_CLIENT_SECRET). Endpoint, resource, and client id default
to the TDS cloud values and can be overridden in config.yaml or via flags.

Examples:
  hlp token                                  # full JSON response
  hlp token | jq -r .accessToken             # bare JWT
  curl -H "Authorization: Bearer $(hlp token | jq -r .accessToken)" ...
  hlp token --resource other-api             # different resource`,
	Args: cobra.NoArgs,
	RunE: runToken,
}

func init() {
	TokenCmd.Flags().StringVar(&flagURL, "url", "", "OAuth endpoint URL (default from config)")
	TokenCmd.Flags().StringVar(&flagResource, "resource", "", "resource to authenticate against (default from config)")
	TokenCmd.Flags().StringVar(&flagClientID, "client-id", "", "OAuth client id (default from config)")
}

// resolveParam returns the flag value when set, otherwise the config value
// (which config.Default() guarantees is populated).
func resolveParam(flagVal, cfgVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return cfgVal
}

// resolveClientSecret returns the OAuth client secret from the environment.
// Precedence: HLP_TDS_CLIENT_SECRET, then TDS_CLIENT_SECRET.
func resolveClientSecret() (string, error) {
	if s := os.Getenv("HLP_TDS_CLIENT_SECRET"); s != "" {
		return s, nil
	}
	if s := os.Getenv("TDS_CLIENT_SECRET"); s != "" {
		return s, nil
	}
	return "", fmt.Errorf("client secret not set: export HLP_TDS_CLIENT_SECRET, " +
		"or use the hlp-token alias which injects it from Infisical (/hlp path)")
}

func runToken(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	secret, err := resolveClientSecret()
	if err != nil {
		return err
	}

	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	resp, err := client.R().
		SetFormData(map[string]string{
			"resource":      resolveParam(flagResource, cfg.Token.Resource),
			"client_id":     resolveParam(flagClientID, cfg.Token.ClientID),
			"client_secret": secret,
		}).
		Post(resolveParam(flagURL, cfg.Token.OAuthURL))
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("token request returned %s: %s", resp.Status(), resp.String())
	}

	fmt.Println(resp.String())
	return nil
}
