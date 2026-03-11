package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func (r *RootCmd) statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "status",
		Short:        "Show agent status",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := r.loadConfig()
			if err != nil {
				return err
			}

			dir := cfg.Dir()
			if dir == "" {
				var e error
				dir, e = configDir()
				if e != nil {
					return e
				}
			}

			// Check if daemon is running via PID file.
			pid, err := readPID(dir)
			if err != nil {
				pid = 0
			}

			connState := "disconnected"
			if isRunning(pid) {
				connState = "connected"
			}

			// Discover active sessions using the default claude dir.
			home, _ := homeDir()
			claudeDir := filepath.Join(home, ".claude")
			sessions := discoverSessions(claudeDir)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Connection:  %s\n", connState)
			fmt.Fprintf(out, "Relay URL:   %s\n", cfg.RelayURL)
			fmt.Fprintf(out, "Machine ID:  %s\n", cfg.MachineID)
			fmt.Fprintf(out, "Sessions:    %d active\n", len(sessions))
			fmt.Fprintf(out, "Latency:     n/a (daemon not queried)\n")
			return nil
		},
	}
}
