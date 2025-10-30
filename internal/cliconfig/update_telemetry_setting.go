package cliconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/example"
	"go.yaml.in/yaml/v3"
)

// UpdateTelemetrySettings updates the telemetry configuration in the config file.
// Uses yaml.Node to preserve comments and formatting.
func UpdateTelemetrySettings(enabled bool) error {
	return UpdateTelemetrySettingsWithOptions(enabled, UpdateOptions{
		GetConfigFile: viper.ConfigFileUsed,
		GetHomeDir:    os.UserHomeDir,
		MkdirAll:      os.MkdirAll,
		GetTemplate:   getExampleTemplate,
	})
}

// UpdateOptions contains dependencies for UpdateTelemetrySettings to enable testing
type UpdateOptions struct {
	GetConfigFile func() string
	GetHomeDir    func() (string, error)
	MkdirAll      func(path string, perm os.FileMode) error
	GetTemplate   func() ([]byte, error)
}

// UpdateTelemetrySettingsWithOptions is the testable version of UpdateTelemetrySettings
func UpdateTelemetrySettingsWithOptions(enabled bool, opts UpdateOptions) error {
	configFile := opts.GetConfigFile()
	if configFile == "" {
		// If no config file exists, create one in default location
		homeDir, err := opts.GetHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".config", "onekeymap")
		configFile = filepath.Join(configDir, "config.yaml")

		// Ensure directory exists
		const configDirMode = 0750
		if err := opts.MkdirAll(configDir, configDirMode); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Get template data
		templateData, err := opts.GetTemplate()
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		// Create initial config based on example template
		return createInitialConfigFileFromTemplate(configFile, enabled, templateData)
	}

	return updateTelemetryInYAMLFile(configFile, enabled)
}

// getExampleTemplate gets the embedded example template
func getExampleTemplate() ([]byte, error) {
	return example.ConfigFS.ReadFile("config.yaml")
}

// createInitialConfigFileFromTemplate creates a config file from a template byte slice
func createInitialConfigFileFromTemplate(configFile string, enabled bool, templateData []byte) error {
	// Parse the template as yaml.Node to preserve comments
	var root yaml.Node
	if err := yaml.Unmarshal(templateData, &root); err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Update the telemetry.enabled value in the template
	updated, err := updateTelemetryEnabledInNode(&root, enabled)
	if err != nil {
		return fmt.Errorf("failed to update telemetry setting in template: %w", err)
	}

	if !updated {
		// If telemetry section doesn't exist in template, add it
		if err := addTelemetrySection(&root, enabled); err != nil {
			return fmt.Errorf("failed to add telemetry section to template: %w", err)
		}
	}

	// Write the updated config to the target file
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	const yamlIndent = 2
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(&root); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	const configFileMode = 0600
	return os.WriteFile(configFile, []byte(buf.String()), configFileMode)
}

// updateTelemetryInYAMLFile updates telemetry.enabled in a YAML file while preserving comments and formatting.
func updateTelemetryInYAMLFile(configFile string, enabled bool) error {
	// Read the current file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse as yaml.Node to preserve comments
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Find and update telemetry.enabled
	updated, err := updateTelemetryEnabledInNode(&root, enabled)
	if err != nil {
		return fmt.Errorf("failed to update telemetry setting: %w", err)
	}

	if !updated {
		// If telemetry section doesn't exist, add it
		if err := addTelemetrySection(&root, enabled); err != nil {
			return fmt.Errorf("failed to add telemetry section: %w", err)
		}
	}

	// Write back to file
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	const yamlIndent = 2
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(&root); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	const configFileMode = 0600
	return os.WriteFile(configFile, []byte(buf.String()), configFileMode)
}

// updateTelemetryEnabledInNode recursively searches for telemetry.enabled and updates it.
// This function also handles uncomment the enabled line if it's commented out.
func updateTelemetryEnabledInNode(node *yaml.Node, enabled bool) (bool, error) {
	if node == nil {
		return false, nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if updated, err := updateTelemetryEnabledInNode(child, enabled); err != nil {
				return false, err
			} else if updated {
				return true, nil
			}
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "telemetry" && valueNode.Kind == yaml.MappingNode {
				// Found telemetry section, look for enabled key
				for j := 0; j < len(valueNode.Content); j += 2 {
					enabledKeyNode := valueNode.Content[j]
					enabledValueNode := valueNode.Content[j+1]

					if enabledKeyNode.Value == "enabled" {
						// Update the enabled value
						enabledValueNode.Value = strconv.FormatBool(enabled)
						enabledValueNode.Tag = "!!bool"
						return true, nil
					}
				}

				// If enabled key doesn't exist in telemetry section, add it
				// Insert it at the beginning of telemetry section
				enabledKeyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "enabled",
				}
				enabledValueNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: strconv.FormatBool(enabled),
					Tag:   "!!bool",
				}
				// Insert at the beginning
				const additionalNodes = 2
				newContent := make([]*yaml.Node, 0, len(valueNode.Content)+additionalNodes)
				newContent = append(newContent, enabledKeyNode, enabledValueNode)
				newContent = append(newContent, valueNode.Content...)
				valueNode.Content = newContent
				return true, nil
			}

			// Recurse into nested structures
			if updated, err := updateTelemetryEnabledInNode(valueNode, enabled); err != nil {
				return false, err
			} else if updated {
				return true, nil
			}
		}
	}

	return false, nil
}

// addTelemetrySection adds a new telemetry section to the root document.
func addTelemetrySection(root *yaml.Node, enabled bool) error {
	// Handle empty document (all comments case)
	if root.Kind != yaml.DocumentNode {
		// Create a proper document structure
		*root = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind:    yaml.MappingNode,
					Content: []*yaml.Node{},
				},
			},
		}
	} else if len(root.Content) == 0 {
		// Document exists but has no content, add a mapping node
		root.Content = []*yaml.Node{
			{
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{},
			},
		}
	}

	docNode := root.Content[0]
	if docNode.Kind != yaml.MappingNode {
		return errors.New("root node is not a mapping")
	}

	// Create telemetry section
	telemetryKeyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "telemetry",
	}

	telemetryValueNode := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "enabled",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: strconv.FormatBool(enabled),
				Tag:   "!!bool",
			},
		},
	}

	// Add to root mapping
	docNode.Content = append(docNode.Content, telemetryKeyNode, telemetryValueNode)

	return nil
}
