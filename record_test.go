package dtpatch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildPatchRecordGolden(t *testing.T) {
	bak, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(bak) = %v", err)
	}
	patched, err := os.ReadFile("testdata/bundle_database.data")
	if err != nil {
		t.Fatalf("ReadFile(patched) = %v", err)
	}

	offBak, err := locateRecord(bak)
	if err != nil {
		t.Fatalf("locateRecord(bak) = %v", err)
	}
	if offBak != 0x122e4c {
		t.Errorf("locateRecord(bak) offset = %#x, want %#x", offBak, 0x122e4c)
	}
	offData, err := locateRecord(patched)
	if err != nil {
		t.Fatalf("locateRecord(patched) = %v", err)
	}

	origRecord := bak[offBak : offBak+OldRecordSize]
	want := patched[offData : offData+184]

	got, err := buildPatchRecord(origRecord, 999)
	if err != nil {
		t.Fatalf("buildPatchRecord() = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("buildPatchRecord() = %x, want %x", got, want)
	}
}

func TestBuildPatchRecordSizeError(t *testing.T) {
	_, err := buildPatchRecord(make([]byte, 80), 999)
	if !errors.Is(err, ErrRecordSize) {
		t.Errorf("buildPatchRecord(80 bytes) error = %v, want %v", err, ErrRecordSize)
	}
}

func TestBuildPatchRecordTruncated(t *testing.T) {
	bak, err := os.ReadFile("testdata/bundle_database.data.bak")
	if err != nil {
		t.Fatalf("ReadFile(bak) = %v", err)
	}
	offBak, err := locateRecord(bak)
	if err != nil {
		t.Fatalf("locateRecord(bak) = %v", err)
	}
	origRecord := append([]byte{}, bak[offBak:offBak+OldRecordSize]...)
	binary.LittleEndian.PutUint32(origRecord[16:20], 0xFFFFFFFF)

	if _, _, err := readString(origRecord, 16); !errors.Is(err, ErrTruncated) {
		t.Errorf("readString() error = %v, want %v", err, ErrTruncated)
	}
	if _, err := buildPatchRecord(origRecord, 999); !errors.Is(err, ErrTruncated) {
		t.Errorf("buildPatchRecord() error = %v, want %v", err, ErrTruncated)
	}
}

func TestStringRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "ascii", in: "9ba626afa44a3aa3"},
		{name: "with dot and digits", in: "9ba626afa44a3aa3.patch_999"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := appendString(nil, c.in)
			got, next, err := readString(buf, 0)
			if err != nil {
				t.Fatalf("readString() = %v", err)
			}
			if diff := cmp.Diff(c.in, got); diff != "" {
				t.Errorf("readString() mismatch (-want +got):\n%s", diff)
			}
			if next != len(buf) {
				t.Errorf("readString() next = %d, want %d", next, len(buf))
			}
		})
	}
}
