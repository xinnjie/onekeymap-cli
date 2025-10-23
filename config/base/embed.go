package base

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"
)

//go:embed *.json
var FS embed.FS

func Read(name string) ([]byte, error) {
	if name == "" {
		return nil, errors.New("base name is empty")
	}
	path := fmt.Sprintf("%s.json", name)
	b, err := FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read base %q: %w", name, err)
	}
	return b, nil
}

func List() ([]string, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		out = append(out, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(out)
	return out, nil
}
