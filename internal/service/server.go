package service

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type Server struct {
	keymapv1.UnimplementedOnekeymapServiceServer
	importer importapi.Importer
	exporter exportapi.Exporter
	registry *plugins.Registry
}

func NewServer(registry *plugins.Registry, importer importapi.Importer, exporter exportapi.Exporter) *Server {
	return &Server{
		importer: importer,
		exporter: exporter,
		registry: registry,
	}
}
