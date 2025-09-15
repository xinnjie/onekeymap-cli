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

type serveFlags struct {
	port   int
	listen string
}

func NewCmdServe() *cobra.Command {
	f := serveFlags{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the gRPC server",
		Run:   serveRun(&f),
		Args:  cobra.ExactArgs(0),
	}

	cmd.Flags().IntVarP(&f.port, "port", "p", 50051, "Port to listen on")
	cmd.Flags().
		StringVar(&f.listen, "listen", "", "Listen address, e.g., tcp://127.0.0.1:50051 or unix:///tmp/onekeymap.sock")

	// Bind listen flag to config key server.listen
	_ = viper.BindPFlag("server.listen", cmd.Flags().Lookup("listen"))

	return cmd
}

func serveRun(f *serveFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// Prefer explicit listen address from config/flag; fallback to --port
		addr := viper.GetString("server.listen")
		if addr == "" {
			addr = fmt.Sprintf(":%d", f.port)
		}
		ctx := cmd.Context()
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
			lc := net.ListenConfig{}
			lis, err = lc.Listen(ctx, "unix", socketPath)
			if err != nil {
				logger.ErrorContext(ctx, "failed to listen on unix socket", "path", socketPath, "err", err)
				os.Exit(1)
			}
			// Best-effort restrict permissions
			_ = os.Chmod(socketPath, 0o600)
			// Ensure cleanup on exit
			defer func() {
				_ = os.Remove(socketPath)
			}()
			logger.InfoContext(ctx, "server listening", "address", "unix://"+socketPath)
		} else {
			// TCP (allow optional tcp:// prefix)
			if after, ok := strings.CutPrefix(addr, "tcp://"); ok {
				addr = after
			}

			if addr == "" {
				logger.ErrorContext(ctx, "empty listen address")
				os.Exit(1)
			}

			lc := net.ListenConfig{}
			lis, err = lc.Listen(ctx, "tcp", addr)
			if err != nil {
				logger.ErrorContext(ctx, "failed to listen on tcp", "address", addr, "err", err)
				os.Exit(1)
			}
			logger.InfoContext(ctx, "server listening", "address", "tcp://"+lis.Addr().String())
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
			logger.ErrorContext(ctx, "failed to serve", "err", err)
			os.Exit(1)
		}
	}
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
