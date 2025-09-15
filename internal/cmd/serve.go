package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
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

		logEvents := func() []grpc_logging.LoggableEvent {
			verbose := viper.GetBool("verbose")
			quiet := viper.GetBool("quiet")

			if quiet {
				return []grpc_logging.LoggableEvent{}
			}

			if verbose {
				return []grpc_logging.LoggableEvent{
					grpc_logging.StartCall,
					grpc_logging.FinishCall,
					grpc_logging.PayloadReceived,
					grpc_logging.PayloadSent,
				}
			}
			return []grpc_logging.LoggableEvent{
				grpc_logging.StartCall,
				grpc_logging.FinishCall,
			}
		}()

		// gRPC logging interceptors (unary and stream)
		opt := []grpc_logging.Option{
			grpc_logging.WithLevels(grpc_logging.DefaultServerCodeToLevel),
			grpc_logging.WithLogOnEvents(logEvents...),
		}

		s := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpc_logging.UnaryServerInterceptor(grpcLogger(logger), opt...),
			),
			grpc.ChainStreamInterceptor(
				grpc_logging.StreamServerInterceptor(grpcLogger(logger), opt...),
			),
		)
		keymapv1.RegisterOnekeymapServiceServer(
			s,
			service.NewServer(pluginRegistry, importService, exportService, mappingConfig, logger),
		)
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

// grpcLogger adapts slog to grpc_logging.Logger.
func grpcLogger(l *slog.Logger) grpc_logging.Logger {
	return grpc_logging.LoggerFunc(func(ctx context.Context, lvl grpc_logging.Level, msg string, fields ...any) {
		switch lvl {
		case grpc_logging.LevelDebug:
			l.Debug(msg, fields...)
		case grpc_logging.LevelInfo:
			l.Info(msg, fields...)
		case grpc_logging.LevelWarn:
			l.Warn(msg, fields...)
		case grpc_logging.LevelError:
			l.Error(msg, fields...)
		default:
			l.Info(msg, fields...)
		}
	})
}
