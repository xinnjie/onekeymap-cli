package vscode

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

func TestVscodeImporter_FindByVSCodeActionWithArgs(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	i := &vscodeLikeImporter{
		mappingConfig: mappingConfig,
		logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	tests := []struct {
		name             string
		command          string
		when             string
		args             map[string]interface{}
		expectedActionID string
		expectedFound    bool
	}{
		{
			name:             "Find with exact match on command, when, and args",
			command:          "cursorEnd",
			when:             "editorTextFocus",
			args:             map[string]interface{}{"sticky": false},
			expectedActionID: "actions.test.withArgs",
			expectedFound:    true,
		},
		{
			name:             "Find with command only, when other when clauses exist",
			command:          "cursorEnd",
			when:             "someOtherWhen",
			args:             map[string]interface{}{"sticky": false},
			expectedActionID: "actions.test.withArgs",
			expectedFound:    true,
		},
		{
			name:             "Find with command and when, no args",
			command:          "editor.action.clipboardCopyAction",
			when:             "editorTextFocus",
			args:             nil,
			expectedActionID: "actions.edit.copy",
			expectedFound:    true,
		},
		{
			name:             "Not found with wrong args",
			command:          "cursorEnd",
			when:             "editorTextFocus",
			args:             map[string]interface{}{"sticky": true},
			expectedActionID: "",
			expectedFound:    false,
		},
		{
			name:             "Not found with wrong command",
			command:          "wrong.command",
			when:             "editorTextFocus",
			args:             nil,
			expectedActionID: "",
			expectedFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := i.FindByVSCodeActionWithArgs(tt.command, tt.when, tt.args)
			if tt.expectedFound {
				assert.NotNil(t, mapping)
				assert.Equal(t, tt.expectedActionID, mapping.ID)
			} else {
				assert.Nil(t, mapping)
			}
		})
	}
}
