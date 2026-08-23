package dtpatch

import (
	"errors"
	"os"
	"testing"
)

func TestLocateRecord(t *testing.T) {
	bak, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(bak) = %v", err)
	}
	patched, err := os.ReadFile("testdata/bundle_database.data")
	if err != nil {
		t.Fatalf("ReadFile(patched) = %v", err)
	}

	if off, err := locateRecord(bak); err != nil || off != 0x122e4c {
		t.Errorf("locateRecord(bak) = (%#x, %v), want (%#x, nil)", off, err, 0x122e4c)
	}
	if off, err := locateRecord(patched); err != nil || off != 0x122e4c {
		t.Errorf("locateRecord(patched) = (%#x, %v), want (%#x, nil)", off, err, 0x122e4c)
	}

	if _, err := locateRecord([]byte("no anchor here")); !errors.Is(err, ErrNotFound) {
		t.Errorf("locateRecord(no anchor) error = %v, want %v", err, ErrNotFound)
	}
}

func TestHasTag(t *testing.T) {
	bak, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(bak) = %v", err)
	}
	patched, err := os.ReadFile("testdata/bundle_database.data")
	if err != nil {
		t.Fatalf("ReadFile(patched) = %v", err)
	}

	if !hasTag(patched, 999) {
		t.Error("hasTag(patched, 999) = false, want true")
	}
	if hasTag(bak, 999) {
		t.Error("hasTag(bak, 999) = true, want false")
	}
	if hasTag(patched, 998) {
		t.Error("hasTag(patched, 998) = true, want false")
	}
}

func TestHasNextPatchWarning(t *testing.T) {
	bak, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(bak) = %v", err)
	}
	patched, err := os.ReadFile("testdata/bundle_database.data")
	if err != nil {
		t.Fatalf("ReadFile(patched) = %v", err)
	}

	if hasNextPatchWarning(bak) {
		t.Error("hasNextPatchWarning(bak) = true, want false")
	}
	if hasNextPatchWarning(patched) {
		t.Error("hasNextPatchWarning(patched) = true, want false")
	}
	if !hasNextPatchWarning([]byte("...9ba626afa44a3aa3.patch_001...")) {
		t.Error("hasNextPatchWarning(synthetic) = false, want true")
	}
}
