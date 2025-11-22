//go:build ignore
// +build ignore

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_selector "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/internal/service"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

const defaultGRPCPort = 50051

type serveFlags struct {
	port   int
	listen string
}

func NewCmdServe(rootFlags *rootFlags) *cobra.Command {
	f := serveFlags{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the gRPC server",
		Run: serveRun(
			&f,
			rootFlags,
			func() (*slog.Logger, *plugins.Registry, importerapi.Importer, exporterapi.Exporter, *mappings.MappingConfig) {
				return cmdLogger, cmdPluginRegistry, cmdImportService, cmdExportService, cmdMappingConfig
			},
		),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().IntVarP(&f.port, "port", "p", defaultGRPCPort, "Port to listen on")
	cmd.Flags().
		StringVar(&f.listen, "listen", "", "Listen address, e.g., tcp://127.0.0.1:50051 or unix:///tmp/onekeymap.sock")

	// Bind listen flag to config key server.listen
	_ = viper.BindPFlag("server.listen", cmd.Flags().Lookup("listen"))

	return cmd
}

func serveRun(
	f *serveFlags,
	rootFlags *rootFlags,
	dependencies func() (*slog.Logger, *plugins.Registry, importerapi.Importer, exporterapi.Exporter, *mappings.MappingConfig),
) func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger, pluginRegistry, importService, exportService, mappingConfig := dependencies()
		// Prefer explicit listen address from config/flag; fallback to --port
		addr := viper.GetString("server.listen")
		sandbox := rootFlags.sandbox
		verbose := rootFlags.verbose
		quiet := rootFlags.quiet
		if addr == "" {
			addr = fmt.Sprintf(":%d", f.port)
		}
		ctx := cmd.Context()
		var (
			lis        net.Listener
			err        error
			socketPath string
		)

		lis, socketPath, err = setupListener(ctx, addr, logger)
		if err != nil {
			os.Exit(1)
		}
		if socketPath != "" {
			defer func() {
				_ = os.Remove(socketPath)
			}()
		}

		logEvents := func() []grpc_logging.LoggableEvent {
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
				grpc_selector.UnaryServerInterceptor(
					grpc_logging.UnaryServerInterceptor(grpcLogger(logger), opt...),
					grpc_selector.MatchFunc(skipHealthCheckRequests),
				),
			),
			grpc.ChainStreamInterceptor(
				grpc_selector.StreamServerInterceptor(
					grpc_logging.StreamServerInterceptor(grpcLogger(logger), opt...),
					grpc_selector.MatchFunc(skipHealthCheckRequests),
				),
			),
		)
		keymapv1.RegisterOnekeymapServiceServer(
			s,
			service.NewServer(
				pluginRegistry,
				importService,
				exportService,
				mappingConfig,
				logger,
				service.ServerOption{Sandbox: sandbox},
			),
		)

		// Register gRPC health check service
		healthServer := health.NewServer()
		healthServer.SetServingStatus("keymap.v1.OnekeymapService", healthv1.HealthCheckResponse_SERVING)
		healthv1.RegisterHealthServer(s, healthServer)

		if err := s.Serve(lis); err != nil {
			logger.ErrorContext(ctx, "failed to serve", "err", err)
			os.Exit(1)
		}
	}
}

func setupListener(ctx context.Context, addr string, logger *slog.Logger) (net.Listener, string, error) {
	if strings.HasPrefix(addr, "unix://") || strings.HasPrefix(addr, "unix:") {
		return setupUnixSocketListener(ctx, addr, logger)
	}
	return setupTCPListener(ctx, addr, logger)
}

func setupUnixSocketListener(ctx context.Context, addr string, logger *slog.Logger) (net.Listener, string, error) {
	socketPath := addr
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
	lis, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		logger.ErrorContext(ctx, "failed to listen on unix socket", "path", socketPath, "err", err)
		return nil, "", err
	}

	// Best-effort restrict permissions
	_ = os.Chmod(socketPath, 0o600)
	logger.InfoContext(ctx, "server listening", "address", "unix://"+socketPath)

	return lis, socketPath, nil
}

func setupTCPListener(ctx context.Context, addr string, logger *slog.Logger) (net.Listener, string, error) {
	// TCP (allow optional tcp:// prefix)
	if after, ok := strings.CutPrefix(addr, "tcp://"); ok {
		addr = after
	}

	if addr == "" {
		logger.ErrorContext(ctx, "empty listen address")
		return nil, "", errors.New("empty listen address")
	}

	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		logger.ErrorContext(ctx, "failed to listen on tcp", "address", addr, "err", err)
		return nil, "", err
	}
	logger.InfoContext(ctx, "server listening", "address", "tcp://"+lis.Addr().String())

	return lis, "", nil
}

// skipHealthCheckRequests filters out health check requests from logging.
func skipHealthCheckRequests(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != healthv1.Health_Check_FullMethodName
}

// grpcLogger adapts slog to grpc_logging.Logger.
func grpcLogger(l *slog.Logger) grpc_logging.Logger {
	return grpc_logging.LoggerFunc(func(_ context.Context, lvl grpc_logging.Level, msg string, fields ...any) {
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
