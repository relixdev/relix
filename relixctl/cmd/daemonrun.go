package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/relixdev/protocol"
	"github.com/spf13/cobra"

	"github.com/relixdev/relix/relixctl/internal/adapter"
	"github.com/relixdev/relix/relixctl/internal/config"
	"github.com/relixdev/relix/relixctl/internal/crypto"
	"github.com/relixdev/relix/relixctl/internal/daemon"
	"github.com/relixdev/relix/relixctl/internal/relay"
)

func (r *RootCmd) daemonRunCmd() *cobra.Command {
	var cfgDirFlag string
	cmd := &cobra.Command{
		Use:          "daemon-run",
		Short:        "Run the relay agent daemon (internal use)",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := cfgDirFlag
			if dir == "" {
				dir = r.cfgDir
			}
			if dir == "" {
				var err error
				dir, err = configDir()
				if err != nil {
					return err
				}
			}
			return runDaemon(dir)
		},
	}
	cmd.Flags().StringVar(&cfgDirFlag, "config-dir", "", "config directory (default ~/.relixctl)")
	return cmd
}

func runDaemon(dir string) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	keys, err := crypto.Load(dir)
	if err != nil {
		// Generate keys on first run.
		keys, err = crypto.GenerateAndSave(dir)
		if err != nil {
			return err
		}
	}

	// Peer public key: in a real deployment this would come from a pairing
	// step. For now use own public key as placeholder so the daemon can start.
	peerPub := keys.PublicKey

	relayClient := relay.NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := relayClient.Connect(ctx, cfg.RelayURL, cfg.AuthToken, cfg.MachineID); err != nil {
		log.Printf("relay connect: %v (will continue offline)", err)
	}

	b := relay.NewBridge(relayClient, keys.PublicKey, keys.PrivateKey, peerPub)

	home, _ := os.UserHomeDir()
	claudeDir := home + "/.claude"
	claudeAdp := adapter.NewClaudeCodeAdapter(claudeDir, nil)
	aiderAdp := adapter.NewAiderAdapter(nil, nil)
	adp := adapter.NewMultiAdapter(claudeAdp, aiderAdp)

	bridgeAdapter := &daemonBridgeAdapter{
		bridge:    b,
		client:    relayClient,
		machineID: cfg.MachineID,
	}
	d := daemon.New(adp, bridgeAdapter, daemon.Options{})

	// Handle OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	return d.Run(ctx)
}

// daemonBridgeAdapter adapts relay.Bridge + relay.RelayClient to daemon.Bridge.
type daemonBridgeAdapter struct {
	bridge    *relay.Bridge
	client    *relay.RelayClient
	machineID string
}

func (a *daemonBridgeAdapter) SendEvent(ctx context.Context, event protocol.Event, sessionID string) error {
	return a.bridge.SendEvent(ctx, event, sessionID)
}

func (a *daemonBridgeAdapter) ReceiveLoop(ctx context.Context) <-chan protocol.Payload {
	return a.bridge.ReceiveLoop(ctx)
}

func (a *daemonBridgeAdapter) SendMachineStatus(ctx context.Context, status string) error {
	return relay.SendMachineStatus(ctx, a.client, a.machineID, status)
}
