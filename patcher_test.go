package dtpatch

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupTempDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	data, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(fixture) = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, BundleDBName), data, 0o644); err != nil {
		t.Fatalf("WriteFile(setup) = %v", err)
	}
	return dir
}

func TestPatchMatchesGolden(t *testing.T) {
	dir := setupTempDB(t)

	if err := Patch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Patch() = %v", err)
	}

	got, err := readBundleDB(dir)
	if err != nil {
		t.Fatalf("readBundleDB() = %v", err)
	}
	want, err := os.ReadFile("testdata/bundle_database.data")
	if err != nil {
		t.Fatalf("ReadFile(golden) = %v", err)
	}

	offGot, err := locateRecord(got)
	if err != nil {
		t.Fatalf("locateRecord(got) = %v", err)
	}
	offWant, err := locateRecord(want)
	if err != nil {
		t.Fatalf("locateRecord(want) = %v", err)
	}

	gotRecord := got[offGot : offGot+184]
	wantRecord := want[offWant : offWant+184]
	if !bytes.Equal(gotRecord, wantRecord) {
		t.Errorf("patched record = %x, want %x", gotRecord, wantRecord)
	}
}

func TestPatchAlreadyPatched(t *testing.T) {
	dir := setupTempDB(t)

	if err := Patch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Patch() = %v", err)
	}
	if err := Patch(Options{Dir: dir, PatchNum: 999}); !errors.Is(err, ErrAlreadyPatched) {
		t.Errorf("Patch() (re-patch) error = %v, want %v", err, ErrAlreadyPatched)
	}
}

func TestUnpatchWithoutBackup(t *testing.T) {
	dir := setupTempDB(t)

	if err := Unpatch(Options{Dir: dir, PatchNum: 999}); !errors.Is(err, ErrNotPatched) {
		t.Errorf("Unpatch() error = %v, want %v", err, ErrNotPatched)
	}
}

func TestUnpatchRestoresExactly(t *testing.T) {
	dir := setupTempDB(t)

	original, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(fixture) = %v", err)
	}

	if err := Patch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Patch() = %v", err)
	}
	if err := Unpatch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Unpatch() = %v", err)
	}

	restored, err := readBundleDB(dir)
	if err != nil {
		t.Fatalf("readBundleDB() = %v", err)
	}
	if !bytes.Equal(restored, original) {
		t.Error("restored bundle database does not match original fixture")
	}
}

func TestToggle(t *testing.T) {
	dir := setupTempDB(t)

	if err := Toggle(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Toggle() (patch) = %v", err)
	}
	patched, err := Status(dir, 999)
	if err != nil {
		t.Fatalf("Status() = %v", err)
	}
	if !patched {
		t.Error("Status() after first Toggle() = false, want true")
	}

	if err := Toggle(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Toggle() (unpatch) = %v", err)
	}
	patched, err = Status(dir, 999)
	if err != nil {
		t.Fatalf("Status() = %v", err)
	}
	if patched {
		t.Error("Status() after second Toggle() = true, want false")
	}
}

func TestPatchBackupBeforeWrite(t *testing.T) {
	dir := setupTempDB(t)

	original, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(fixture) = %v", err)
	}

	if err := Patch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Patch() = %v", err)
	}

	backup, err := os.ReadFile(filepath.Join(dir, BackupName))
	if err != nil {
		t.Fatalf("ReadFile(backup) = %v", err)
	}
	if !bytes.Equal(backup, original) {
		t.Error("backup does not match pre-patch original")
	}
}

func TestStatus(t *testing.T) {
	dir := setupTempDB(t)

	patched, err := Status(dir, 999)
	if err != nil {
		t.Fatalf("Status() (before) = %v", err)
	}
	if patched {
		t.Error("Status() before Patch() = true, want false")
	}

	if err := Patch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Patch() = %v", err)
	}
	patched, err = Status(dir, 999)
	if err != nil {
		t.Fatalf("Status() (after patch) = %v", err)
	}
	if !patched {
		t.Error("Status() after Patch() = false, want true")
	}

	if err := Unpatch(Options{Dir: dir, PatchNum: 999}); err != nil {
		t.Fatalf("Unpatch() = %v", err)
	}
	patched, err = Status(dir, 999)
	if err != nil {
		t.Fatalf("Status() (after unpatch) = %v", err)
	}
	if patched {
		t.Error("Status() after Unpatch() = true, want false")
	}
}
