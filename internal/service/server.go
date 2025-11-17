package service

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type ServerOption struct {
	Sandbox bool
}

type Server struct {
	keymapv1.UnimplementedOnekeymapServiceServer

	importer      importerapi.Importer
	exporter      exporterapi.Exporter
	registry      *plugins.Registry
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger

	opt ServerOption
}

func NewServer(
	registry *plugins.Registry,
	importer importerapi.Importer,
	exporter exporterapi.Exporter,
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
