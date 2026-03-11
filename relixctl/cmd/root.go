package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/relixdev/relix/relixctl/internal/config"
)

const version = "0.1.0"

// RootCmd wraps cobra.Command so callers can set output for testing.
type RootCmd struct {
	*cobra.Command
	cfgDir string
}

// New returns a configured RootCmd. cfgDir overrides the default ~/.relixctl/
// and is used in tests to isolate config state.
func New(cfgDir string) *RootCmd {
	root := &RootCmd{cfgDir: cfgDir}

	rootCmd := &cobra.Command{
		Use:     "relixctl",
		Short:   "relixctl — Relix agent CLI",
		Version: version,
		// Silence usage on error so tests don't flood stdout.
		SilenceUsage: true,
	}

	rootCmd.AddCommand(root.configCmd())
	rootCmd.AddCommand(stubCmd("login", "Authenticate with Relix"))
	rootCmd.AddCommand(stubCmd("pair", "Pair a mobile device"))
	rootCmd.AddCommand(root.statusCmd())
	rootCmd.AddCommand(root.sessionsCmd())
	rootCmd.AddCommand(root.startCmd())
	rootCmd.AddCommand(root.stopCmd())
	rootCmd.AddCommand(root.uninstallCmd())
	rootCmd.AddCommand(root.daemonRunCmd())

	root.Command = rootCmd
	return root
}

func stubCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:          use,
		Short:        short,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: not yet implemented\n", use)
			return nil
		},
	}
}

func (r *RootCmd) loadConfig() (*config.Config, error) {
	if r.cfgDir != "" {
		return config.Load(r.cfgDir)
	}
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	return config.Load(dir)
}

func (r *RootCmd) configCmd() *cobra.Command {
	cfgCmd := &cobra.Command{
		Use:          "config",
		Short:        "Manage relixctl configuration",
		SilenceUsage: true,
	}
	cfgCmd.AddCommand(r.configSetCmd())
	cfgCmd.AddCommand(r.configGetCmd())
	return cfgCmd
}

var validKeys = map[string]bool{
	"relay_url":  true,
	"auth_token": true,
	"machine_id": true,
}

func (r *RootCmd) configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "set <key> <value>",
		Short:        "Set a config value",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, val := args[0], args[1]
			if !validKeys[key] {
				return fmt.Errorf("unknown config key %q", key)
			}
			cfg, err := r.loadConfig()
			if err != nil {
				return err
			}
			switch key {
			case "relay_url":
				cfg.RelayURL = val
			case "auth_token":
				cfg.AuthToken = val
			case "machine_id":
				cfg.MachineID = val
			}
			return cfg.Save()
		},
	}
}

func (r *RootCmd) configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get <key>",
		Short:        "Get a config value",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !validKeys[key] {
				return fmt.Errorf("unknown config key %q", key)
			}
			cfg, err := r.loadConfig()
			if err != nil {
				return err
			}
			var val string
			switch key {
			case "relay_url":
				val = cfg.RelayURL
			case "auth_token":
				val = cfg.AuthToken
			case "machine_id":
				val = cfg.MachineID
			}
			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		},
	}
}
