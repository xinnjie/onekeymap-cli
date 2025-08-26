package mappings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateCheckWithArgs(t *testing.T) {
	testCases := []struct {
		name                  string
		yamlContent           string
		expectError           bool
		expectedMappingsCount int
	}{
		{
			name: "VSCode: same command different args should not be duplicate",
			yamlContent: `
mappings:
  - id: "test.action.1"
    description: "Test action 1"
    vscode:
      command: "cursorEnd"
      args:
        "sticky": false
  - id: "test.action.2"
    description: "Test action 2"
    vscode:
      command: "cursorEnd"
      args:
        "sticky": true
`,
			expectError:           false,
			expectedMappingsCount: 2,
		},
		{
			name: "VSCode: same command same args should be duplicate",
			yamlContent: `
mappings:
  - id: "test.action.1"
    description: "Test action 1"
    vscode:
      command: "cursorEnd"
      args:
        "sticky": false
  - id: "test.action.2"
    description: "Test action 2"
    vscode:
      command: "cursorEnd"
      args:
        "sticky": false
`,
			expectError: true,
		},
		{
			name: "Zed: same action different args should not be duplicate",
			yamlContent: `
mappings:
  - id: "test.action.1"
    description: "Test action 1"
    zed:
      action: "editor::MoveToEnd"
      args:
        "extend": false
      context: "Editor"
  - id: "test.action.2"
    description: "Test action 2"
    zed:
      action: "editor::MoveToEnd"
      args:
        "extend": true
      context: "Editor"
`,
			expectError:           false,
			expectedMappingsCount: 2,
		},
		{
			name: "Zed: same action same args should be duplicate",
			yamlContent: `
mappings:
  - id: "test.action.1"
    description: "Test action 1"
    zed:
      action: "editor::MoveToEnd"
      args:
        "extend": false
      context: "Editor"
  - id: "test.action.2"
    description: "Test action 2"
    zed:
      action: "editor::MoveToEnd"
      args:
        "extend": false
      context: "Editor"
`,
			expectError: true,
		},
		{
			name: "No args vs empty args should be different",
			yamlContent: `
mappings:
  - id: "test.action.1"
    description: "Test action 1"
    vscode:
      command: "cursorEnd"
  - id: "test.action.2"
    description: "Test action 2"
    vscode:
      command: "cursorEnd"
      args: {}
`,
			expectError:           false,
			expectedMappingsCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.yamlContent)
			config, err := load(reader)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tc.expectedMappingsCount, len(config.Mappings))
			}
		})
	}
}
