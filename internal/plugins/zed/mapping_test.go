package zed

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestActionIDFromZedWithArgs(t *testing.T) {
	mappingConfig, _ := mappings.NewTestMappingConfig()
	zedImporter := newImporter(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())

	tests := []struct {
		name          string
		action        string
		context       string
		args          map[string]interface{}
		expectedID    string
		expectError   bool
		expectedError string
	}{
		{
			name:       "Exact match with args",
			action:     "test::ActionWithArgs",
			context:    "Editor",
			args:       map[string]interface{}{"test_param": true},
			expectedID: "actions.test.withArgs",
		},
		{
			name:       "Match without args",
			action:     "editor::Copy",
			context:    "Editor",
			args:       nil,
			expectedID: "actions.edit.copy",
		},
		{
			name:          "No match with wrong args",
			action:        "test::ActionWithArgs",
			context:       "Editor",
			args:          map[string]interface{}{"test_param": false},
			expectError:   true,
			expectedError: "no mapping found for zed action: test::ActionWithArgs",
		},
		{
			name:          "No match with wrong action",
			action:        "wrong::action",
			context:       "Test",
			args:          map[string]interface{}{"foo": "bar"},
			expectError:   true,
			expectedError: "no mapping found for zed action: wrong::action",
		},
		{
			name:          "No match with wrong context",
			action:        "test::ActionWithArgs",
			context:       "WrongContext",
			args:          map[string]interface{}{"test_param": true},
			expectError:   true,
			expectedError: "no mapping found for zed action: test::ActionWithArgs",
		},
		{
			name:          "Args provided but mapping has none",
			action:        "editor::Copy",
			context:       "Editor",
			args:          map[string]interface{}{"foo": "bar"},
			expectError:   true,
			expectedError: "no mapping found for zed action: editor::Copy",
		},
		{
			name:          "Args missing but mapping has them",
			action:        "test::ActionWithArgs",
			context:       "Test",
			args:          nil,
			expectError:   true,
			expectedError: "no mapping found for zed action: test::ActionWithArgs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualID, err := zedImporter.actionIDFromZedWithArgs(tt.action, tt.context, tt.args)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, actualID)
			}
		})
	}
}
