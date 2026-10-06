package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

// Regressão (#57): organizar → desfazer não remove pastas que já existiam no
// destino, mesmo vazias; as que a organização criou são removidas.
func TestOrganizeThenUndoKeepsExistingFolders(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dst, "txt")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	st := store.New(filepath.Join(t.TempDir(), "last-operation.json"), nil)
	req := organizer.Request{
		SourceFolderPath:      src,
		DestinationFolderPath: dst,
		Options:               criteria.Options{ByExtension: true, BySize: true},
	}

	if _, err := organizer.NewService(organizer.Deps{Metadata: metadata.Reader{}, Store: st}).Organize(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(existing, "size-1mb", "a.txt")); err != nil {
		t.Fatalf("o arquivo deveria estar em txt/size-1mb: %v", err)
	}
	if _, err := undo.NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(existing); err != nil {
		t.Fatalf("a pasta txt/ já existia e deveria continuar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(existing, "size-1mb")); !os.IsNotExist(err) {
		t.Fatalf("a pasta criada size-1mb/ deveria ter sido removida (stat err = %v)", err)
	}
}
