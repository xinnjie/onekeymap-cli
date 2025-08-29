package service

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type Server struct {
	keymapv1.UnimplementedOnekeymapServiceServer
	importer importapi.Importer
	exporter exportapi.Exporter
}

func NewServer(importer importapi.Importer, exporter exportapi.Exporter) *Server {
	return &Server{
		importer: importer,
		exporter: exporter,
	}
}
