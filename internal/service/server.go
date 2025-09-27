package service

import (
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type ServerOption struct {
	Sandbox bool
}

type Server struct {
	keymapv1.UnimplementedOnekeymapServiceServer

	importer      importapi.Importer
	exporter      exportapi.Exporter
	registry      *plugins.Registry
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger

	opt ServerOption
}

func NewServer(
	registry *plugins.Registry,
	importer importapi.Importer,
	exporter exportapi.Exporter,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	opt ServerOption,
) *Server {
	return &Server{
		importer:      importer,
		exporter:      exporter,
		registry:      registry,
		mappingConfig: mappingConfig,
		logger:        logger,
		opt:           opt,
	}
}
