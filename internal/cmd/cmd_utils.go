package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func confirm(cmd *cobra.Command, path string) bool {
	if path == "" {
		panic("path is empty")
	}
	cmd.Printf("Write config to %s? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func backupIfExists(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if !info.Mode().IsRegular() {
		// Only back up regular files
		return "", nil
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ts := time.Now().Format("20060102-150405")
	backup := filepath.Join(dir, base+".bak-"+ts)
	// Ensure uniqueness if a backup with the same timestamp exists
	for i := 1; ; i++ {
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			break
		}
		backup = filepath.Join(dir, fmt.Sprintf("%s.bak-%s-%d", base, ts, i))
	}

	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(backup, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return backup, nil
}
