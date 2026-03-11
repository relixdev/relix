package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func (r *RootCmd) sessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "sessions",
		Short:        "List active Claude Code sessions",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := homeDir()
			if err != nil {
				return err
			}
			claudeDir := filepath.Join(home, ".claude")
			sessions := discoverSessions(claudeDir)

			out := cmd.OutOrStdout()
			// Table header.
			fmt.Fprintf(out, "%-36s  %-30s  %-10s  %s\n", "ID", "Project", "Status", "Started")
			fmt.Fprintf(out, "%-36s  %-30s  %-10s  %s\n",
				"------------------------------------",
				"------------------------------",
				"----------",
				"-------------------",
			)

			if len(sessions) == 0 {
				fmt.Fprintln(out, "No active sessions found.")
				return nil
			}

			for _, s := range sessions {
				project := s.Project
				if len(project) > 30 {
					project = "..." + project[len(project)-27:]
				}
				started := time.Unix(s.StartedAt, 0).Format("2006-01-02 15:04:05")
				fmt.Fprintf(out, "%-36s  %-30s  %-10s  %s\n",
					s.ID, project, string(s.Status), started)
			}
			return nil
		},
	}
}
