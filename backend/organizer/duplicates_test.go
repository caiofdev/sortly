package organizer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/backup"
	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Origem com nota.txt e o destino já com txt/nota.txt (#82).
func duplicateSetup(t *testing.T) (src, dst string) {
	t.Helper()
	src, dst = t.TempDir(), t.TempDir()
	writeAt(t, filepath.Join(src, "nota.txt"), "nova", testTime)
	writeAt(t, filepath.Join(dst, "txt", "nota.txt"), "antiga", testTime)
	return src, dst
}

func plannedMove(src, dst string) Plan {
	return Plan{Destination: dst, Moves: []Move{{filepath.Join(src, "nota.txt"), filepath.Join(dst, "txt", "nota.txt")}}}
}

func TestValidDuplicates(t *testing.T) {
	for policy, want := range map[string]bool{"rename": true, "skip": true, "replace": true, "": false, "merge": false} {
		if got := ValidDuplicates(policy); got != want {
			t.Errorf("ValidDuplicates(%q) = %v, want %v", policy, got, want)
		}
	}
}

func TestApplyDuplicates(t *testing.T) {
	tests := []struct {
		name         string
		policy       string
		backup       bool
		wantTree     []string
		wantSkipped  int
		wantReplaced int
	}{
		{"renomear (padrão da 1.0)", DuplicatesRename, false, []string{"txt/nota (1).txt", "txt/nota.txt"}, 0, 0},
		{"política vazia renomeia", "", false, []string{"txt/nota (1).txt", "txt/nota.txt"}, 0, 0},
		{"ignorar deixa na origem", DuplicatesSkip, false, []string{"txt/nota.txt"}, 1, 0},
		{"substituir guarda o existente", DuplicatesReplace, true, []string{"txt/nota.txt"}, 0, 1},
		{"substituir sem backup renomeia", DuplicatesReplace, false, []string{"txt/nota (1).txt", "txt/nota.txt"}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := duplicateSetup(t)
			plan := plannedMove(src, dst)
			plan.Duplicates = tt.policy
			if tt.backup {
				plan.BackupFolder = filepath.Join(t.TempDir(), "substituidos", "1")
			}

			out, err := NewExecutor(nil).Apply(context.Background(), plan)

			if err != nil || out.SkippedDuplicates != tt.wantSkipped || len(out.ReplacedItems) != tt.wantReplaced {
				t.Fatalf("Apply = (%+v, %v)", out, err)
			}
			assertTree(t, dst, tt.wantTree)
			checkDuplicateFiles(t, src, plan, out)
		})
	}
}

// O ignorado continua na origem; o substituído foi para o backup, na mesma
// posição relativa ao destino (#82).
func checkDuplicateFiles(t *testing.T, src string, plan Plan, out Outcome) {
	t.Helper()
	if out.SkippedDuplicates > 0 && readFile(t, filepath.Join(src, "nota.txt")) != "nova" {
		t.Fatal("ignorado deveria ficar na origem")
	}
	for _, item := range out.ReplacedItems {
		if readFile(t, item.Path) != "nova" || readFile(t, item.Backup) != "antiga" ||
			item.Backup != filepath.Join(plan.BackupFolder, "txt", "nota.txt") {
			t.Fatalf("substituição = %+v", item)
		}
	}
}

// Dois arquivos desta organização com o mesmo nome indo para a mesma pasta: o
// segundo é renomeado, nunca ignorado nem substituindo o primeiro (#82).
func TestDuplicatePolicyOnlyForPreexistingFiles(t *testing.T) {
	for _, policy := range []string{DuplicatesSkip, DuplicatesReplace} {
		t.Run(policy, func(t *testing.T) {
			a, b, dst := t.TempDir(), t.TempDir(), t.TempDir()
			writeAt(t, filepath.Join(a, "x.txt"), "a", testTime)
			writeAt(t, filepath.Join(b, "x.txt"), "b", testTime)
			plan := Plan{Destination: dst, Duplicates: policy, BackupFolder: filepath.Join(t.TempDir(), "1"), Moves: []Move{
				{filepath.Join(a, "x.txt"), filepath.Join(dst, "x.txt")},
				{filepath.Join(b, "x.txt"), filepath.Join(dst, "x.txt")},
			}}

			out, err := NewExecutor(nil).Apply(context.Background(), plan)

			if err != nil || len(out.MovedItems) != 2 || out.SkippedDuplicates != 0 || len(out.ReplacedItems) != 0 {
				t.Fatalf("Apply = (%+v, %v)", out, err)
			}
			assertTree(t, dst, []string{"x (1).txt", "x.txt"})
		})
	}
}

func replacePlan(t *testing.T) (src, dst string, plan Plan) {
	t.Helper()
	src, dst = duplicateSetup(t)
	plan = plannedMove(src, dst)
	plan.Duplicates, plan.BackupFolder = DuplicatesReplace, filepath.Join(t.TempDir(), "1")
	return src, dst, plan
}

func TestReplaceBackupFolderFails(t *testing.T) {
	_, dst, plan := replacePlan(t)
	e := NewExecutor(nil)
	e.mkdirAll = func(dir string) error {
		if filepath.Base(dir) == "txt" && filepath.Dir(dir) == plan.BackupFolder {
			return os.ErrPermission
		}
		return os.MkdirAll(dir, 0o755)
	}

	out, _ := e.Apply(context.Background(), plan)

	if out.FailedFiles != 1 || readFile(t, filepath.Join(dst, "txt", "nota.txt")) != "antiga" {
		t.Fatalf("out = %+v", out)
	}
}

func TestReplaceNewFileFailsRestoresExisting(t *testing.T) {
	src, dst, plan := replacePlan(t)
	e := NewExecutor(nil)
	realMove := e.move
	e.move = func(from, to string) (string, error) {
		if from == filepath.Join(src, "nota.txt") {
			return "", os.ErrPermission
		}
		return realMove(from, to)
	}

	out, _ := e.Apply(context.Background(), plan)

	if out.FailedFiles != 1 || len(out.ReplacedItems) != 0 || readFile(t, filepath.Join(dst, "txt", "nota.txt")) != "antiga" {
		t.Fatalf("out = %+v", out)
	}
	if readFile(t, filepath.Join(src, "nota.txt")) != "nova" {
		t.Fatal("o novo deveria continuar na origem")
	}
}

func TestReplaceRestoreFailsKeepsBackup(t *testing.T) {
	src, _, plan := replacePlan(t)
	e := NewExecutor(nil)
	realMove := e.move
	e.move = func(from, to string) (string, error) {
		if from == filepath.Join(src, "nota.txt") || filepath.Dir(filepath.Dir(from)) == plan.BackupFolder {
			return "", os.ErrPermission
		}
		return realMove(from, to)
	}

	out, _ := e.Apply(context.Background(), plan)

	if out.FailedFiles != 1 || readFile(t, filepath.Join(plan.BackupFolder, "txt", "nota.txt")) != "antiga" {
		t.Fatalf("out = %+v", out)
	}
}

func TestApplyReportingCountsSkipped(t *testing.T) {
	src, dst := duplicateSetup(t)
	plan := plannedMove(src, dst)
	plan.Duplicates = DuplicatesSkip
	var last Progress
	out, err := NewExecutor(nil).ApplyReporting(context.Background(), plan, func(p Progress) { last = p })
	if err != nil || out.SkippedDuplicates != 1 || last != (Progress{Done: 1, Total: 1}) {
		t.Fatalf("ApplyReporting = (%+v, %v), último progresso %+v", out, err, last)
	}
}

type fakeBackups struct {
	folder  string
	removed []string
}

func (f *fakeBackups) NewFolder() string    { return f.folder }
func (f *fakeBackups) Remove(folder string) { f.removed = append(f.removed, folder) }

func TestOrganizeReplaceRecordsBackup(t *testing.T) {
	src, dst := duplicateSetup(t)
	st := store.New(filepath.Join(t.TempDir(), ".sortly", "last-operation.json"), nil)
	backups := &fakeBackups{folder: filepath.Join(t.TempDir(), "substituidos", "1")}
	if err := st.Save(store.Operation{SourceFolderPath: "x", MovedItems: []store.MovedItem{{From: "a", To: "b"}}, BackupFolder: "anterior"}); err != nil {
		t.Fatal(err)
	}
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: st, Location: testLoc, Backups: backups})

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, DestinationFolderPath: dst,
		Options: criteria.Default, Duplicates: DuplicatesReplace})

	op, _ := st.Load()
	if err != nil || got.ReplacedFiles != 1 || len(op.ReplacedItems) != 1 || op.BackupFolder != backups.folder {
		t.Fatalf("Organize = (%+v, %v), registro = %+v", got, err, op)
	}
	if len(backups.removed) != 1 || backups.removed[0] != "anterior" {
		t.Fatalf("o backup do registro substituído deveria ser apagado: %v", backups.removed)
	}
}

func TestOrganizeSkipCountsAndKeepsRecordClean(t *testing.T) {
	src, dst := duplicateSetup(t)
	writeAt(t, filepath.Join(src, "outro.txt"), "x", testTime)
	svc, st := newTestService(t)

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, DestinationFolderPath: dst,
		Options: criteria.Default, Duplicates: DuplicatesSkip})

	op, _ := st.Load()
	if err != nil || got.SkippedDuplicates != 1 || got.MovedFiles != 1 || len(op.ReplacedItems) != 0 || op.BackupFolder != "" {
		t.Fatalf("Organize = (%+v, %v), registro = %+v", got, err, op)
	}
}

func TestOrganizeUnknownPolicyRenames(t *testing.T) {
	src, dst := duplicateSetup(t)
	svc, _ := newTestService(t)
	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, DestinationFolderPath: dst,
		Options: criteria.Default, Duplicates: "mesclar"})
	if err != nil || got.MovedFiles != 1 {
		t.Fatalf("Organize = (%+v, %v)", got, err)
	}
	assertTree(t, dst, []string{"txt/nota (1).txt", "txt/nota.txt"})
}

func TestRecordNotSavedRemovesPreviousBackup(t *testing.T) {
	src, dst := duplicateSetup(t)
	backups := &fakeBackups{}
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: &brokenStore{previous: &store.Operation{BackupFolder: "anterior"}}, Location: testLoc, Backups: backups})

	_, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, DestinationFolderPath: dst, Options: criteria.Default})

	if !errors.Is(err, ErrRecordNotSaved) || len(backups.removed) != 1 || backups.removed[0] != "anterior" {
		t.Fatalf("err = %v, apagados = %v", err, backups.removed)
	}
}

type brokenStore struct{ previous *store.Operation }

func (b *brokenStore) Load() (*store.Operation, error) { return b.previous, nil }
func (b *brokenStore) Save(store.Operation) error      { return os.ErrPermission }
func (b *brokenStore) Clear() error                    { return nil }

func TestBackupStoreSatisfiesInterface(t *testing.T) {
	var _ BackupStore = backup.New(t.TempDir(), nil)
}
