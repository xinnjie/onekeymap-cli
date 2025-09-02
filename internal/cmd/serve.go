package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/service"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		// Prefer explicit listen address from config/flag; fallback to --port
		addr := viper.GetString("server.listen")
		if addr == "" {
			port, _ := cmd.Flags().GetInt("port")
			addr = fmt.Sprintf(":%d", port)
		}

		var (
			lis        net.Listener
			err        error
			socketPath string
		)

		if strings.HasPrefix(addr, "unix://") || strings.HasPrefix(addr, "unix:") {
			// Unix domain socket
			socketPath = addr
			if strings.HasPrefix(socketPath, "unix://") {
				socketPath = strings.TrimPrefix(socketPath, "unix://")
			} else {
				socketPath = strings.TrimPrefix(socketPath, "unix:")
			}
			// Remove stale file if exists
			if _, statErr := os.Stat(socketPath); statErr == nil {
				_ = os.Remove(socketPath)
			}
			lis, err = net.Listen("unix", socketPath)
			if err != nil {
				logger.Error("failed to listen on unix socket", "path", socketPath, "err", err.Error())
				os.Exit(1)
			}
			// Best-effort restrict permissions
			_ = os.Chmod(socketPath, 0o600)
			// Ensure cleanup on exit
			defer func() {
				_ = os.Remove(socketPath)
			}()
			logger.Info("server listening", "address", "unix://"+socketPath)
		} else {
			// TCP (allow optional tcp:// prefix)
			if after, ok := strings.CutPrefix(addr, "tcp://"); ok {
				addr = after
			}

			if addr == "" {
				logger.Error("empty listen address")
				os.Exit(1)
			}

			lis, err = net.Listen("tcp", addr)
			if err != nil {
				logger.Error("failed to listen on tcp", "address", addr, "err", err.Error())
				os.Exit(1)
			}
			logger.Info("server listening", "address", "tcp://"+lis.Addr().String())
		}

		s := grpc.NewServer()
		keymapv1.RegisterOnekeymapServiceServer(s, service.NewServer(pluginRegistry, importService, exportService, mappingConfig, logger))
		if err := s.Serve(lis); err != nil {
			logger.Error("failed to serve", "err", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 50051, "Port to listen on")
	serveCmd.Flags().String("listen", "", "Listen address, e.g., tcp://127.0.0.1:50051 or unix:///tmp/onekeymap.sock")
	// Bind listen flag to config key server.listen
	_ = viper.BindPFlag("server.listen", serveCmd.Flags().Lookup("listen"))
}
