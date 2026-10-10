package undo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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

// Os movidos e os substituídos saem numa barra só: o total soma os dois planos, e a
// pasta é relativa a quem recebe o arquivo (a origem, depois o destino) (#96).
func TestUndoReportingProgress(t *testing.T) {
	st, backups, _, _, _ := replacedSetup(t)
	var got []organizer.Progress

	res, err := NewService(st, nil).WithBackups(backups).UndoReporting(context.Background(), func(p organizer.Progress) {
		got = append(got, p)
	})

	if err != nil || res.RestoredFiles != 1 || res.RestoredReplaced != 1 {
		t.Fatalf("UndoReporting = (%+v, %v)", res, err)
	}
	want := []organizer.Progress{
		{Done: 0, Total: 2, File: "nota.txt", Folder: ""},
		{Done: 1, Total: 2},
		{Done: 1, Total: 2, File: "nota.txt", Folder: "txt"},
		{Done: 2, Total: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("progresso =\n%+v\nwant\n%+v", got, want)
	}
}

func TestUndoReportingCases(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name         string
		ctx          context.Context
		record       bool
		missing      bool
		wantErr      error
		wantCanceled bool
		wantProgress []organizer.Progress
	}{
		{"sem registro, nada a reportar", context.Background(), false, false, ErrNothingToUndo, false, nil},
		{"cancelado antes do 1º arquivo: tudo fica no registro", canceled, true, false, context.Canceled, true, nil},
		{"arquivo sumiu: só o fim, 0 de 0", context.Background(), true, true, nil, false, []organizer.Progress{{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := t.TempDir(), t.TempDir()
			st := newStore(t)
			if tt.record {
				item := move(t, src, dst, "a.txt", "txt")
				if tt.missing {
					if err := os.Remove(item.To); err != nil {
						t.Fatal(err)
					}
				}
				saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{item}})
			}
			var got []organizer.Progress

			res, err := NewService(st, nil).UndoReporting(tt.ctx, func(p organizer.Progress) { got = append(got, p) })

			if !errors.Is(err, tt.wantErr) || res.Canceled != tt.wantCanceled || !reflect.DeepEqual(got, tt.wantProgress) {
				t.Fatalf("UndoReporting = (%+v, %v), progresso %+v", res, err, got)
			}
			if tt.wantCanceled && !res.CanUndo {
				t.Fatal("cancelado, o desfazer continua disponível")
			}
		})
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

func (f *failingReplacedMover) ApplyReporting(ctx context.Context, plan organizer.Plan, report organizer.Reporter) (organizer.Outcome, error) {
	if len(plan.Moves) > 0 && filepath.Dir(filepath.Dir(plan.Moves[0].From)) == f.backupFolder {
		return organizer.Outcome{FailedFiles: len(plan.Moves)}, nil
	}
	return f.real.ApplyReporting(ctx, plan, report)
}
