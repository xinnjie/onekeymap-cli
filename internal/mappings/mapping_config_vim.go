package mappings

type VimMappingConfig struct {
	EditorActionMapping `yaml:",inline"`
	Command             string `yaml:"command"`
	Mode                string `yaml:"mode"`
}

func checkVimDuplicateConfig(_ map[string]ActionMappingConfig) error {
	// TODO(xinnjie): Vim plugin not implemented
	return nil
}
