package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/relixdev/relix/relixctl/internal/auth"
	"github.com/relixdev/relix/relixctl/internal/crypto"
)

const (
	defaultClientID      = "relix-cli"
	defaultAuthURL       = "https://github.com/login/oauth/authorize"
	defaultTokenURL      = "https://api.relix.sh/auth/github"
	defaultDeviceCodeURL = "https://github.com/login/device/code"
)

// loginOpts holds injectable dependencies for the login command.
type loginOpts struct {
	browserLogin func(clientID, authURL, tokenURL string) (string, error)
	deviceLogin  func(clientID, deviceCodeURL, tokenURL string) (string, error)
}

func (r *RootCmd) loginCmd() *cobra.Command {
	var headless bool

	cmd := &cobra.Command{
		Use:          "login",
		Short:        "Authenticate with Relix",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve functions: test override > default.
			opts := loginOpts{
				browserLogin: auth.BrowserLogin,
				deviceLogin:  auth.DeviceCodeLogin,
			}
			if r.overrideLoginOpts != nil {
				if r.overrideLoginOpts.browserLogin != nil {
					opts.browserLogin = r.overrideLoginOpts.browserLogin
				}
				if r.overrideLoginOpts.deviceLogin != nil {
					opts.deviceLogin = r.overrideLoginOpts.deviceLogin
				}
			}

			cfg, err := r.loadConfig()
			if err != nil {
				return err
			}

			var token string
			if headless {
				token, err = opts.deviceLogin(defaultClientID, defaultDeviceCodeURL, defaultTokenURL)
			} else {
				token, err = opts.browserLogin(defaultClientID, defaultAuthURL, defaultTokenURL)
			}
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Store token in config.
			cfg.AuthToken = token
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			// Generate X25519 keypair if none exists.
			keysDir := filepath.Join(cfg.Dir(), "keys")
			if _, err := os.Stat(filepath.Join(keysDir, "public.key")); os.IsNotExist(err) {
				if _, err := crypto.GenerateAndSave(keysDir); err != nil {
					return fmt.Errorf("generate keypair: %w", err)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Logged in as authenticated user")
			return nil
		},
	}

	cmd.Flags().BoolVar(&headless, "headless", false, "Use device code flow (for headless environments)")
	return cmd
}
