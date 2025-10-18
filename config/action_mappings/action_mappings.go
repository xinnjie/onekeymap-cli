package actionmappings

import (
	"embed"
	"io"
	"io/fs"
	"sort"
	"strings"
)

var (
	//go:embed *.yaml
	actionMappingsFS embed.FS
)

func ReadActionMapping() (io.Reader, error) {
	names, _ := fs.Glob(actionMappingsFS, "*.yaml")
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasSuffix(name, "test.yaml") {
			continue
		}
		data, err := actionMappingsFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		parts = append(parts, string(data))
	}
	allYamls := strings.Join(parts, "\n---\n")
	return strings.NewReader(allYamls), nil
}

func ReadTestActionMapping() (io.Reader, error) {
	names, _ := fs.Glob(actionMappingsFS, "*test.yaml")
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		data, err := actionMappingsFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		parts = append(parts, string(data))
	}
	allYamls := strings.Join(parts, "\n---\n")
	return strings.NewReader(allYamls), nil
}
