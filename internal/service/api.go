package service

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ExportKeymap(
	ctx context.Context,
	req *keymapv1.ExportKeymapRequest,
) (*keymapv1.ExportKeymapResponse, error) {
	// Validate editor type and ensure it's supported
	if req.GetEditorType() == keymapv1.EditorType_EDITOR_TYPE_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "editor_type is required")
	}
	et := pluginapi.EditorType(strings.ToLower(req.GetEditorType().String()))
	if _, ok := s.registry.Get(et); !ok {
		return nil, status.Errorf(codes.NotFound, "editor not supported: %s", et)
	}

	// Only pass Base when provided; if empty, do not pass
	var base io.Reader
	if strings.TrimSpace(req.GetBase()) != "" {
		base = strings.NewReader(req.GetBase())
	}

	var buf bytes.Buffer
	report, err := s.exporter.Export(ctx, &buf, req.GetKeymap(), exportapi.ExportOptions{
		EditorType: et,
		Base:       base,
		DiffType:   req.GetDiffType(),
		FilePath:   req.GetFilePath(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %s", err)
	}

	return &keymapv1.ExportKeymapResponse{
		Keymap: buf.String(),
		Diff:   report.Diff,
	}, nil
}

func (s *Server) ImportKeymap(
	ctx context.Context,
	req *keymapv1.ImportKeymapRequest,
) (*keymapv1.ImportKeymapResponse, error) {
	// Validate editor type and ensure it's supported
	if req.GetEditorType() == keymapv1.EditorType_EDITOR_TYPE_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "editor_type is required")
	}
	et := pluginapi.EditorType(strings.ToLower(req.GetEditorType().String()))
	if _, ok := s.registry.Get(et); !ok {
		return nil, status.Errorf(codes.NotFound, "editor not supported: %s", et)
	}

	var baseSetting *keymapv1.Keymap
	if req.GetBase() != "" {
		km, err := keymap.Load(strings.NewReader(req.GetBase()))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse base keymap: %v", err)
		}
		baseSetting = km
	}

	result, err := s.importer.Import(ctx, importapi.ImportOptions{
		EditorType:  et,
		InputStream: strings.NewReader(req.GetSource()),
		Base:        baseSetting,
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

func toProtoKeymapDiff(diffs []importapi.KeymapDiff) []*keymapv1.ActionDiff {
	var result []*keymapv1.ActionDiff
	for _, d := range diffs {
		result = append(result, &keymapv1.ActionDiff{
			Origin:  d.Before,
			Updated: d.After,
		})
	}
	return result
}

func (s *Server) ConfigDetect(
	ctx context.Context,
	req *keymapv1.ConfigDetectRequest,
) (*keymapv1.ConfigDetectResponse, error) {
	// For now, plugins resolve path by runtime.GOOS. We only support macOS requests on this server currently.
	switch req.GetPlatform() {
	case keymapv1.Platform_MACOS:
		et := pluginapi.EditorType(strings.ToLower(req.GetEditorType().String()))
		plugin, ok := s.registry.Get(et)
		if !ok {
			return nil, status.Errorf(codes.NotFound, "editor not supported: %s", et)
		}
		v, _, err := plugin.ConfigDetect(pluginapi.ConfigDetectOptions{Sandbox: s.opt.Sandbox})
		if err != nil || len(v) == 0 {
			return nil, status.Errorf(codes.NotFound, "no default config paths found for editor: %s", et)
		}
		return &keymapv1.ConfigDetectResponse{Paths: v}, nil
	default:
		return nil, status.Errorf(codes.Unimplemented, "platform %s not supported yet", req.GetPlatform().String())
	}
}

func (s *Server) LoadKeymap(
	ctx context.Context,
	req *keymapv1.GetKeymapRequest,
) (*keymapv1.GetKeymapResponse, error) {
	km, err := keymap.Load(strings.NewReader(req.GetConfig()))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse keymap config: %v", err)
	}

	// If return_all is true, we will return all mappings in the config.
	if req.GetReturnAll() {
		existingBindings := make(map[string]*keymapv1.ActionBinding)
		for _, binding := range km.GetKeybindings() {
			existingBindings[binding.GetId()] = binding
		}

		for id, mapping := range s.mappingConfig.Mappings {
			if _, exists := existingBindings[id]; !exists {
				km.Keybindings = append(km.Keybindings, &keymapv1.ActionBinding{
					Id:          id,
					Description: mapping.Description,
					Name:        mapping.Name,
					Category:    mapping.Category,
				})
			}
		}
	}

	km = keymap.DecorateSetting(km, s.mappingConfig)

	return &keymapv1.GetKeymapResponse{Keymap: km}, nil
}

func (s *Server) SaveKeymap(
	ctx context.Context,
	req *keymapv1.SaveKeymapRequest,
) (*keymapv1.SaveKeymapResponse, error) {
	var buf bytes.Buffer
	err := keymap.Save(&buf, req.GetKeymap())
	if err != nil {
		return nil, err
	}

	return &keymapv1.SaveKeymapResponse{
		Config: buf.String(),
	}, nil
}
