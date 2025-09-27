package intellij

import (
	"os"
	"strings"
)

// isIntelliJInstalled checks if any JetBrains IDE is installed on macOS.
func isIntelliJInstalled() (bool, error) {
	appsDir := "/Applications"
	dentries, err := os.ReadDir(appsDir)
	if err != nil {
		return false, err
	}

	jetbrainsIDEs := []string{
		"IntelliJ IDEA",
		"GoLand",
		"PyCharm",
		"WebStorm",
		"PhpStorm",
		"Rider",
		"CLion",
		"RubyMine",
		"AppCode",
		"DataGrip",
		"DataSpell",
	}

	for _, dentry := range dentries {
		if dentry.IsDir() {
			continue
		}
		if strings.HasSuffix(dentry.Name(), ".app") {
			appName := strings.TrimSuffix(dentry.Name(), ".app")
			for _, ide := range jetbrainsIDEs {
				if strings.HasPrefix(appName, ide) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
