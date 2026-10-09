package tests

// Ponta a ponta com "Substituir": o arquivo que já estava no destino vai para o
// backup e volta no desfazer; o backup some quando o registro é desfeito ou
// substituído por outra organização (ADR 0008, #82).

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/backup"
	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

type replaceEnv struct {
	src, dst string
	st       *store.FileStore
	backups  *backup.Store
	root     string
}

func newReplaceEnv(t *testing.T) replaceEnv {
	t.Helper()
	home := t.TempDir()
	env := replaceEnv{
		src:  t.TempDir(),
		dst:  t.TempDir(),
		st:   store.New(filepath.Join(home, "last-operation.json"), nil),
		root: filepath.Join(home, "substituidos"),
	}
	env.backups = backup.New(env.root, nil)
	write(t, filepath.Join(env.src, "nota.txt"), "nova")
	write(t, filepath.Join(env.dst, "txt", "nota.txt"), "antiga")
	return env
}

func (e replaceEnv) organize(t *testing.T) organizer.Result {
	t.Helper()
	svc := organizer.NewService(organizer.Deps{Metadata: metadata.Reader{}, Store: e.st, Backups: e.backups})
	res, err := svc.Organize(context.Background(), organizer.Request{
		SourceFolderPath: e.src, DestinationFolderPath: e.dst,
		Options: criteria.Default, Duplicates: organizer.DuplicatesReplace,
	})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestReplaceThenUndoRestoresBoth(t *testing.T) {
	env := newReplaceEnv(t)

	res := env.organize(t)
	if res.ReplacedFiles != 1 || read(t, filepath.Join(env.dst, "txt", "nota.txt")) != "nova" {
		t.Fatalf("Organize = %+v", res)
	}
	op, _ := env.st.Load()

	got, err := undo.NewService(env.st, nil).WithBackups(env.backups).Undo(context.Background())

	if err != nil || got.RestoredFiles != 1 || got.RestoredReplaced != 1 {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	if read(t, filepath.Join(env.src, "nota.txt")) != "nova" || read(t, filepath.Join(env.dst, "txt", "nota.txt")) != "antiga" {
		t.Fatal("cada arquivo deveria voltar ao seu lugar")
	}
	if files.Exists(op.BackupFolder) {
		t.Fatal("depois do desfazer, o backup é apagado")
	}
}

func TestNextOrganizationRemovesPreviousBackup(t *testing.T) {
	env := newReplaceEnv(t)
	env.organize(t)
	first, _ := env.st.Load()

	write(t, filepath.Join(env.src, "outra.txt"), "x")
	env.organize(t)

	if files.Exists(first.BackupFolder) {
		t.Fatal("o registro novo substitui o anterior; o backup dele não tem mais como voltar")
	}
	if !filepath.IsAbs(first.BackupFolder) || !env.backups.Contains(first.BackupFolder) {
		t.Fatalf("pasta do backup = %q", first.BackupFolder)
	}
}
