package dtpatch

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	BundleDBName        = "bundle_database.data"
	BackupName          = "bundle_database.data.bak"
	BootBundleNextPatch = "9ba626afa44a3aa3.patch_001"
	OldRecordSize       = 84
)

var startingPoint = [8]byte{0xA3, 0x3A, 0x4A, 0xA4, 0xAF, 0x26, 0xA6, 0x9B}

var ErrNotFound = errors.New("boot bundle record not found")

func locateRecord(data []byte) (int, error) {
	offset := bytes.Index(data, startingPoint[:])
	if offset < 0 {
		return 0, ErrNotFound
	}
	return offset, nil
}

func hasTag(data []byte, patchNum int) bool {
	return bytes.Contains(data, fmt.Appendf(nil, "patch_%d", patchNum))
}

func hasNextPatchWarning(data []byte) bool {
	return bytes.Contains(data, []byte(BootBundleNextPatch))
}

func readBundleDB(dir string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dir, BundleDBName))
}

func writeBundleDB(dir string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filepath.Join(dir, BundleDBName), data, perm)
}

func backupBundleDB(dir string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filepath.Join(dir, BackupName), data, perm)
}

func restoreBundleDB(dir string) error {
	return os.Rename(filepath.Join(dir, BackupName), filepath.Join(dir, BundleDBName))
}
