package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestServer_LoadKeymap(t *testing.T) {
	ctx := context.Background()
	mockMappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"action1": {Description: "Description 1"},
			"action2": {Description: "Description 2"},
			"action3": {Description: "Description 3"},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := NewServer(nil, nil, nil, mockMappingConfig, logger, ServerOption{Sandbox: false})

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
					Keybindings: []*keymapv1.ActionBinding{
						{
							Id:          "action1",
							Description: "Description 1",
							Bindings: []*keymapv1.Binding{
								{
									KeyChords: &keymapv1.KeyChordSequence{
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
					Keybindings: []*keymapv1.ActionBinding{
						{
							Id:          "action1",
							Description: "Description 1",
							Bindings: []*keymapv1.Binding{
								{
									KeyChords: &keymapv1.KeyChordSequence{
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
						{Id: "action2", Description: "Description 2"},
						{Id: "action3", Description: "Description 3"},
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
					Keybindings: []*keymapv1.ActionBinding{
						{Id: "action1", Description: "Description 1"},
						{Id: "action2", Description: "Description 2"},
						{Id: "action3", Description: "Description 3"},
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
			got, err := server.LoadKeymap(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			} else {
				require.NoError(t, err)
				if tt.expectNoOrder {
					assert.Empty(t, cmp.Diff(tt.want, got, protocmp.Transform(), protocmp.SortRepeatedFields(&keymapv1.Keymap{}, "keybindings")))
				} else {
					assert.Empty(t, cmp.Diff(tt.want, got, protocmp.Transform()))
				}
			}
		})
	}
}
