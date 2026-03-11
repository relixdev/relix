package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (r *RootCmd) uninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "uninstall",
		Short:        "Remove relixctl from this machine",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := r.cfgDir
			if dir == "" {
				var err error
				dir, err = configDir()
				if err != nil {
					return err
				}
			}

			// Stop daemon if running.
			pid, err := readPID(dir)
			if err == nil && isRunning(pid) {
				_ = stopDaemon(dir)
			}

			// Remove the config directory (keys, config, PID file, logs).
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("uninstall: remove config dir: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Uninstall complete. Removed %s\n", dir)
			return nil
		},
	}
}
