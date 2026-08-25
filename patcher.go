package dtpatch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultPatchNum = 999

type Options struct {
	Dir      string
	PatchNum int
	Force    bool
}

var (
	ErrAlreadyPatched     = errors.New("database already patched")
	ErrUnexpectedPatch001 = errors.New("unexpected pre-existing patch_001 record")
	ErrNotPatched         = errors.New("database does not appear patched")
)

func resolvePatchNum(n int) int {
	if n == 0 {
		return defaultPatchNum
	}
	return n
}

func Patch(opts Options) error {
	num := resolvePatchNum(opts.PatchNum)

	data, err := readBundleDB(opts.Dir)
	if err != nil {
		return err
	}
	if hasNextPatchWarning(data) {
		return ErrUnexpectedPatch001
	}
	if hasTag(data, num) {
		return ErrAlreadyPatched
	}

	offset, err := locateRecord(data)
	if err != nil {
		return err
	}
	origRecord := data[offset : offset+OldRecordSize]

	newRecord, err := buildPatchRecord(origRecord, num)
	if err != nil {
		return err
	}

	if err := backupBundleDB(opts.Dir, data, 0o644); err != nil {
		return fmt.Errorf("backup bundle database: %w", err)
	}

	patched := make([]byte, 0, len(data)-OldRecordSize+len(newRecord))
	patched = append(patched, data[:offset]...)
	patched = append(patched, newRecord...)
	patched = append(patched, data[offset+OldRecordSize:]...)

	return writeBundleDB(opts.Dir, patched, 0o644)
}

func Unpatch(opts Options) error {
	num := resolvePatchNum(opts.PatchNum)

	data, err := readBundleDB(opts.Dir)
	if err != nil {
		return err
	}
	if !opts.Force && !hasTag(data, num) {
		return ErrNotPatched
	}

	backupPath := filepath.Join(opts.Dir, BackupName)
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("%w: no backup found: %v", ErrNotPatched, err)
	}

	return restoreBundleDB(opts.Dir)
}

func Toggle(opts Options) error {
	patched, err := Status(opts.Dir, opts.PatchNum)
	if err != nil {
		return err
	}
	if patched {
		return Unpatch(opts)
	}
	return Patch(opts)
}

func Status(dir string, num int) (bool, error) {
	data, err := readBundleDB(dir)
	if err != nil {
		return false, err
	}
	return hasTag(data, resolvePatchNum(num)), nil
}
