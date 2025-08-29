package service

import (
	"bytes"
	"context"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *Server) ExportKeymap(ctx context.Context, req *keymapv1.ExportKeymapRequest) (*keymapv1.ExportKeymapResponse, error) {
	var buf bytes.Buffer
	report, err := s.exporter.Export(&buf, req.Keymap, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType(strings.ToLower(req.EditorType.String())),
		Base:       strings.NewReader(req.Base),
	})
	if err != nil {
		return nil, err
	}

	return &keymapv1.ExportKeymapResponse{
		Keymap: buf.String(),
		Diff:   report.Diff,
	}, nil
}

func (s *Server) ImportKeymap(ctx context.Context, req *keymapv1.ImportKeymapRequest) (*keymapv1.ImportKeymapResponse, error) {

	var baseSetting keymapv1.KeymapSetting
	if req.Base != "" {
		if err := protojson.Unmarshal([]byte(req.Base), &baseSetting); err != nil {
			return nil, err
		}
	}

	result, err := s.importer.Import(ctx, importapi.ImportOptions{
		EditorType:  pluginapi.EditorType(strings.ToLower(req.EditorType.String())),
		InputStream: strings.NewReader(req.Source),
		Base:        &baseSetting,
	})
	if err != nil {
		return nil, err
	}

	return &keymapv1.ImportKeymapResponse{
		Keymap: result.Setting,
		Changes: &keymapv1.KeymapChanges{
			Add:    result.Changes.Add,
			Remove: result.Changes.Remove,
			Update: toProtoKeymapDiff(result.Changes.Update),
		},
	}, nil
}

func toProtoKeymapDiff(diffs []importapi.KeymapDiff) []*keymapv1.KeymapDiff {
	var result []*keymapv1.KeymapDiff
	for _, d := range diffs {
		result = append(result, &keymapv1.KeymapDiff{
			Origin:  d.Before,
			Updated: d.After,
		})
	}
	return result
}
