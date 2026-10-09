package undo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/backup"
	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/store"
)

func content(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Como "Substituir" deixaria: a nota nova em dst/txt e a antiga no backup (#82).
func replacedSetup(t *testing.T) (st *store.FileStore, backups *backup.Store, src, dst, folder string) {
	t.Helper()
	src, dst = t.TempDir(), t.TempDir()
	root := filepath.Join(t.TempDir(), "substituidos")
	backups = backup.New(root, nil)
	folder = filepath.Join(root, "1")
	writeFile(t, filepath.Join(folder, "txt", "nota.txt"), "antiga")
	item := move(t, src, dst, "nota.txt", "txt")
	st = newStore(t)
	op := store.Operation{
		SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{item},
		ReplacedItems: []store.ReplacedItem{{Path: item.To, Backup: filepath.Join(folder, "txt", "nota.txt")}},
		BackupFolder:  folder,
	}
	if err := st.Save(op); err != nil {
		t.Fatal(err)
	}
	return st, backups, src, dst, folder
}

func TestUndoRestoresReplaced(t *testing.T) {
	st, backups, src, dst, folder := replacedSetup(t)

	got, err := NewService(st, nil).WithBackups(backups).Undo(context.Background())

	if err != nil || got.RestoredFiles != 1 || got.RestoredReplaced != 1 || got.CanUndo {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	if content(t, filepath.Join(src, "nota.txt")) != "nota.txt" || content(t, filepath.Join(dst, "txt", "nota.txt")) != "antiga" {
		t.Fatal("o novo volta para a origem e o antigo para o lugar dele")
	}
	if files.Exists(folder) {
		t.Fatal("sem registro, o backup é apagado")
	}
	if op, _ := st.Load(); op != nil {
		t.Fatalf("registro deveria ter sido apagado: %+v", op)
	}
}

func TestUndoReplacedSafety(t *testing.T) {
	tests := []struct {
		name string
		edit func(op *store.Operation, dst string)
	}{
		// Regressão do #54 aplicada aos substituídos: um registro editado não pode
		// mover para fora do destino nem ler de fora do backup (#82).
		{"destino fora das pastas registradas", func(op *store.Operation, dst string) {
			op.ReplacedItems[0].Path = filepath.Join(filepath.Dir(dst), "fora.txt")
		}},
		{"backup fora da raiz dos substituídos", func(op *store.Operation, _ string) {
			op.ReplacedItems[0].Backup = filepath.Join(os.TempDir(), "qualquer.txt")
		}},
		{"backup que não existe mais", func(op *store.Operation, _ string) {
			op.ReplacedItems[0].Backup += ".sumiu"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, backups, _, dst, _ := replacedSetup(t)
			op, _ := st.Load()
			tt.edit(op, dst)
			if err := st.Save(*op); err != nil {
				t.Fatal(err)
			}

			got, err := NewService(st, nil).WithBackups(backups).Undo(context.Background())

			if err != nil || got.RestoredFiles != 1 || got.RestoredReplaced != 0 {
				t.Fatalf("Undo = (%+v, %v)", got, err)
			}
		})
	}
}

func TestUndoWithoutBackupsSkipsReplaced(t *testing.T) {
	st, _, _, dst, folder := replacedSetup(t)

	got, err := NewService(st, nil).Undo(context.Background())

	if err != nil || got.RestoredReplaced != 0 || files.Exists(filepath.Join(dst, "txt", "nota.txt")) {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	if !files.Exists(folder) {
		t.Fatal("sem o Backups configurado, nada fora do desfazer é apagado")
	}
}

func TestUndoReplacedFailureKeepsRecord(t *testing.T) {
	st, backups, _, _, folder := replacedSetup(t)
	svc := NewService(st, nil).WithBackups(backups)
	svc.mover = &failingReplacedMover{real: svc.mover, backupFolder: folder}

	got, err := svc.Undo(context.Background())

	op, _ := st.Load()
	if err != nil || got.RestoredReplaced != 0 || got.FailedFiles != 1 || !got.CanUndo {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	if op == nil || len(op.MovedItems) != 0 || len(op.ReplacedItems) != 1 || !files.Exists(folder) {
		t.Fatalf("o substituído que falhou fica no registro e no backup: %+v", op)
	}
}

func TestReplacedOnlyRecordCanUndo(t *testing.T) {
	op := &store.Operation{ReplacedItems: []store.ReplacedItem{{Path: "a", Backup: "b"}}}
	if !op.CanUndo() {
		t.Fatal("um substituído pendente mantém o desfazer")
	}
}

type failingReplacedMover struct {
	real         Mover
	backupFolder string
}

func (f *failingReplacedMover) Apply(ctx context.Context, plan organizer.Plan) (organizer.Outcome, error) {
	if len(plan.Moves) > 0 && filepath.Dir(filepath.Dir(plan.Moves[0].From)) == f.backupFolder {
		return organizer.Outcome{FailedFiles: len(plan.Moves)}, nil
	}
	return f.real.Apply(ctx, plan)
}
