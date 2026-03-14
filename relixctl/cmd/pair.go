package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/auth"
	"github.com/relixdev/relix/relixctl/internal/crypto"
)

// pairOpts holds injectable dependencies for the pair command.
type pairOpts struct {
	pairFn func(relayURL, authToken, code string, ks *crypto.KeyPair, keysDir string) (protocol.PublicKey, error)
}

func (r *RootCmd) pairCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "pair <code>",
		Short:        "Pair a mobile device",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			code := args[0]

			// Resolve pair function: test override > default.
			pairFn := auth.Pair
			if r.overridePairOpts != nil && r.overridePairOpts.pairFn != nil {
				pairFn = r.overridePairOpts.pairFn
			}

			cfg, err := r.loadConfig()
			if err != nil {
				return err
			}

			if cfg.AuthToken == "" {
				return fmt.Errorf("not logged in — run 'relixctl login' first")
			}

			// Load or generate keypair.
			keysDir := filepath.Join(cfg.Dir(), "keys")
			kp, err := crypto.Load(keysDir)
			if err != nil {
				kp, err = crypto.GenerateAndSave(keysDir)
				if err != nil {
					return fmt.Errorf("generate keypair: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Enter this code in the Relix app: %s\n", code)
			fmt.Fprintln(cmd.OutOrStdout(), "Waiting for mobile device to complete pairing...")

			// Build WebSocket URL for pairing from relay URL.
			pairURL := cfg.RelayURL + "/pair"

			_, err = pairFn(pairURL, cfg.AuthToken, code, kp, keysDir)
			if err != nil {
				return fmt.Errorf("pairing failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Paired with mobile device")
			return nil
		},
	}
}
