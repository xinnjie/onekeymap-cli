//go:build disable

package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/service"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestServer_GetKeymap(t *testing.T) {
	ctx := context.Background()
	mockMappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"action1": {Description: "Description 1"},
			"action2": {Description: "Description 2"},
			"action3": {Description: "Description 3"},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := service.NewServer(nil, nil, nil, mockMappingConfig, logger, service.ServerOption{Sandbox: false})

	tests := []struct {
		name          string
		req           *keymapv1.GetKeymapRequest
		want          *keymapv1.GetKeymapResponse
		wantErrCode   codes.Code
		wantErr       bool
		expectNoOrder bool
	}{
		{
			name: "Invalid config with keys field, ReturnAll=false",
			req: &keymapv1.GetKeymapRequest{
				Config: `{"keybindings":[{"id":"action1","keys":["ctrl+a"]}]}`,
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Valid config, ReturnAll=false",
			req: &keymapv1.GetKeymapRequest{
				Config: `{"version":"1.0","keymaps":[{"id":"action1","keybinding":"ctrl+a"}]}`,
			},
			want: &keymapv1.GetKeymapResponse{
				Keymap: &keymapv1.Keymap{
					Actions: []*keymapv1.Action{
						{
							Name: "action1",
							ActionConfig: &keymapv1.ActionConfig{
								Description: "Description 1",
							},
							Bindings: []*keymapv1.KeybindingReadable{
								{
									KeyChords: &keymapv1.Keybinding{
										Chords: []*keymapv1.KeyChord{
											{
												KeyCode: keymapv1.KeyCode_A,
												Modifiers: []keymapv1.KeyModifier{
													keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
												},
											},
										},
									},
									KeyChordsReadable: "ctrl+a",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Invalid config with keys field, ReturnAll=true",
			req: &keymapv1.GetKeymapRequest{
				Config:    `{"keybindings":[{"id":"action1","keys":["ctrl+a"]}]}`,
				ReturnAll: true,
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Valid config, ReturnAll=true",
			req: &keymapv1.GetKeymapRequest{
				Config:    `{"version":"1.0","keymaps":[{"id":"action1","keybinding":"ctrl+a"}]}`,
				ReturnAll: true,
			},
			want: &keymapv1.GetKeymapResponse{
				Keymap: &keymapv1.Keymap{
					Actions: []*keymapv1.Action{
						{
							Name: "action1",
							ActionConfig: &keymapv1.ActionConfig{
								Description: "Description 1",
							},
							Bindings: []*keymapv1.KeybindingReadable{
								{
									KeyChords: &keymapv1.Keybinding{
										Chords: []*keymapv1.KeyChord{
											{
												KeyCode: keymapv1.KeyCode_A,
												Modifiers: []keymapv1.KeyModifier{
													keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
												},
											},
										},
									},
									KeyChordsReadable: "ctrl+a",
								},
							},
						},
						{Name: "action2", ActionConfig: &keymapv1.ActionConfig{Description: "Description 2"}},
						{Name: "action3", ActionConfig: &keymapv1.ActionConfig{Description: "Description 3"}},
					},
				},
			},
			expectNoOrder: true,
		},
		{
			name: "Empty config, ReturnAll=false",
			req: &keymapv1.GetKeymapRequest{
				Config: "",
			},
			want: &keymapv1.GetKeymapResponse{
				Keymap: &keymapv1.Keymap{},
			},
		},
		{
			name: "Empty config, ReturnAll=true",
			req: &keymapv1.GetKeymapRequest{
				Config:    "",
				ReturnAll: true,
			},
			want: &keymapv1.GetKeymapResponse{
				Keymap: &keymapv1.Keymap{
					Actions: []*keymapv1.Action{
						{Name: "action1", ActionConfig: &keymapv1.ActionConfig{Description: "Description 1"}},
						{Name: "action2", ActionConfig: &keymapv1.ActionConfig{Description: "Description 2"}},
						{Name: "action3", ActionConfig: &keymapv1.ActionConfig{Description: "Description 3"}},
					},
				},
			},
			expectNoOrder: true,
		},
		{
			name: "Invalid config, ReturnAll=false",
			req: &keymapv1.GetKeymapRequest{
				Config: "invalid json",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid config, ReturnAll=true",
			req: &keymapv1.GetKeymapRequest{
				Config:    "invalid json",
				ReturnAll: true,
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := server.GetKeymap(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			} else {
				require.NoError(t, err)
				// Ignore editor_support field in comparison as it's tested separately
				ignoreEditorSupport := protocmp.IgnoreFields(&keymapv1.ActionConfig{}, "editor_support")
				if tt.expectNoOrder {
					assert.Empty(t, cmp.Diff(tt.want, got, protocmp.Transform(), protocmp.SortRepeatedFields(&keymapv1.Keymap{}, "actions"), ignoreEditorSupport))
				} else {
					assert.Empty(t, cmp.Diff(tt.want, got, protocmp.Transform(), ignoreEditorSupport))
				}
			}
		})
	}
}
