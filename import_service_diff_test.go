package onekeymap

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func setting(kms ...*keymapv1.KeyBinding) *keymapv1.KeymapSetting {
	return &keymapv1.KeymapSetting{Keybindings: kms}
}

func newServiceWithImportData(t *testing.T, imported *keymapv1.KeymapSetting) importapi.Importer {
	t.Helper()
	reg := plugins.NewRegistry()
	reg.Register(newTestPlugin(pluginapi.EditorTypeVSCode, "", imported, nil))
	mc, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewImportService(reg, mc, logger, metrics.NewNoop())
}

func TestImportService_Diff_NoBaseline_AllAdd(t *testing.T) {
	imp := setting(
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
		keymap.NewBinding("actions.editor.paste", "cmd+v"),
	)
	service := newServiceWithImportData(t, imp)
	res, err := service.Import(context.Background(), importapi.ImportOptions{EditorType: pluginapi.EditorTypeVSCode, InputStream: strings.NewReader("dummy")})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Changes)
	assert.Equal(t, 2, len(res.Changes.Add))
	assert.Equal(t, 0, len(res.Changes.Remove))
	assert.Equal(t, 0, len(res.Changes.Update))
}

func TestImportService_Diff_Identical_NoChanges(t *testing.T) {
	imp := setting(
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
		keymap.NewBinding("actions.editor.paste", "cmd+v"),
	)
	base := setting(
		keymap.NewBinding("actions.editor.paste", "cmd+v"),
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
	)
	service := newServiceWithImportData(t, imp)
	res, err := service.Import(context.Background(), importapi.ImportOptions{EditorType: pluginapi.EditorTypeVSCode, InputStream: strings.NewReader("dummy"), Base: base})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Changes)
	assert.Equal(t, 0, len(res.Changes.Add))
	assert.Equal(t, 0, len(res.Changes.Remove))
	assert.Equal(t, 0, len(res.Changes.Update))
}

func TestImportService_Diff_AddRemove(t *testing.T) {
	imp := setting(
		keymap.NewBinding("actions.editor.paste", "ctrl+v"),
		keymap.NewBinding("actions.file.save", "ctrl+s"),
	)
	base := setting(
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
		keymap.NewBinding("actions.editor.paste", "ctrl+v"),
	)
	service := newServiceWithImportData(t, imp)
	res, err := service.Import(context.Background(), importapi.ImportOptions{EditorType: pluginapi.EditorTypeVSCode, InputStream: strings.NewReader("dummy"), Base: base})
	require.NoError(t, err)
	require.NotNil(t, res)
	// Add: save; Remove: copy
	require.NotNil(t, res.Changes)
	require.Len(t, res.Changes.Add, 1)
	require.Len(t, res.Changes.Remove, 1)
	assert.Equal(t, "actions.file.save", res.Changes.Add[0].Id)
	assert.Equal(t, "actions.editor.copy", res.Changes.Remove[0].Id)
	assert.Empty(t, res.Changes.Update)
}

func TestImportService_Diff_SingleUpdate(t *testing.T) {
	imp := setting(
		keymap.NewBinding("actions.editor.copy", "cmd+c"),
	)
	base := setting(
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
	)
	service := newServiceWithImportData(t, imp)
	res, err := service.Import(context.Background(), importapi.ImportOptions{EditorType: pluginapi.EditorTypeVSCode, InputStream: strings.NewReader("dummy"), Base: base})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Changes)
	assert.Len(t, res.Changes.Update, 1)
	upd := res.Changes.Update[0]
	require.NotNil(t, upd.Before)
	require.NotNil(t, upd.After)
	// Ensure adds/removes suppressed for the updated pair
	assert.Empty(t, res.Changes.Add)
	assert.Empty(t, res.Changes.Remove)
	// Validate the specific before/after pair
	assert.Equal(t, pairKey(base.Keybindings[0]), pairKey(upd.Before))
	assert.Equal(t, pairKey(imp.Keybindings[0]), pairKey(upd.After))
}

func TestImportService_Diff_MultiPerAction_NoUpdate(t *testing.T) {
	imp := setting(
		keymap.NewBinding("actions.editor.build", "cmd+b"),
	)
	base := setting(
		keymap.NewBinding("actions.editor.build", "ctrl+b"),
		keymap.NewBinding("actions.editor.build", "alt+b"),
	)
	service := newServiceWithImportData(t, imp)
	res, err := service.Import(context.Background(), importapi.ImportOptions{EditorType: pluginapi.EditorTypeVSCode, InputStream: strings.NewReader("dummy"), Base: base})
	require.NoError(t, err)
	require.NotNil(t, res)
	// No update since baseline has 2 entries for the action
	require.NotNil(t, res.Changes)
	assert.Empty(t, res.Changes.Update)
	// Adds contains the new binding; removes contain both baseline bindings
	require.Len(t, res.Changes.Add, 1)
	require.Len(t, res.Changes.Remove, 2)
	assert.Equal(t, "actions.editor.build", res.Changes.Add[0].Id)
	for _, r := range res.Changes.Remove {
		assert.Equal(t, "actions.editor.build", r.Id)
	}
}
