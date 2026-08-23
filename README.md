# dtpatch

`dtpatch` patches Darktide's `bundle_database.data`, enabling
the game to load Darktide Mod Loader (DML) content. It computes the
patch record at runtime using the already known binary layout.

## Contents

- `bundledb.go` file I/O and constants: reading/writing/backing up
  `bundle_database.data`, locating the boot bundle record by raw byte
  search (`locateRecord`), and checking already-patched /
  next-patch-warning state (`hasTag`, `hasNextPatchWarning`).
- `record.go` binary record construction: `buildPatchRecord` builds the
  184-byte patched record from an 84-byte original record and a patch
  number, plus the length-prefixed string helpers (`readString`,
  `appendString`) it depends on.
- `patcher.go` the public API (`Options`, `Patch`, `Unpatch`, `Toggle`,
  `Status`): orchestrates the read -> check -> locate -> build -> backup ->
  write sequence and defines the sentinel errors
  (`ErrAlreadyPatched`, `ErrUnexpectedPatch001`, `ErrNotPatched`)
  callers can `errors.Is` against.
- `cmd/dtpatch.go` the CLI entrypoint: flag parsing for the
  `patch`/`unpatch`/`toggle` subcommands and friendly error-message
  mapping.
- `testdata/` real captured fixtures used as golden data by the test
  suite, see Development below.

## How it works

Stingray/Bitsquid engine stores a binary index of game content
bundles in `bundle_database.data`. To let the game load a mod-supplied
"patch bundle", an entry describing that bundle must be injected into the
database. `dtpatch` does this by locating a known bundle's record, the
game's **boot bundle**, hex id `9ba626afa44a3aa3`, and splicing in new
bytes at that offset.

1. Read `bundle_database.data` into memory.
2. Abort if the database already contains the target patch tag (already
   patched), or if it contains an unexpected pre-existing
   `9ba626afa44a3aa3.patch_001` record (a state this tool doesn't know how
   to handle safely).
3. Search for the boot bundle's hash bytes to find the target record's
   file offset. This is a raw byte search, not a fixed offset: the
   offset shifts across game versions, but the hash bytes do not.
4. Write a backup of the untouched database to `bundle_database.data.bak`
   before modifying anything.
5. Replace the 84-byte record at that offset with a newly built 184-byte
   patched record (layout below).
6. Write the modified buffer back to `bundle_database.data`.

Unpatching simply renames `bundle_database.data.bak` back over
`bundle_database.data`.

### Binary record format

The record format was reverse-engineered by diffing a real unpatched
(`.bak`) and patched (`.data`) database file, byte-verified with `xxd`
against the fixtures in `testdata/`. All offsets are relative to the
record start; the anchor (boot bundle hash) is found by raw substring
search, not a fixed file offset.

**Unpatched record (84 bytes):**

| offset | size | field             | value                                                         |
| ------ | ---- | ----------------- | -------------------------------------------------------------- |
| 0x00   | 8    | hash              | `A3 3A 4A A4 AF 26 A6 9B` (boot bundle hash, little-endian)   |
| 0x08   | 4    | counter (u32 LE)  | 1                                                              |
| 0x0C   | 4    | count (u32 LE)    | 4 (??)                                                         |
| 0x10   | 4    | len1 (u32 LE)     | 16                                                              |
| 0x14   | 16   | str1              | `"9ba626afa44a3aa3"`                                           |
| 0x24   | 4    | len2 (u32 LE)     | 23                                                              |
| 0x28   | 23   | str2              | `"9ba626afa44a3aa3.stream"`                                    |
| 0x3F   | 21   | zero padding      | (out to offset 0x54 = 84 bytes total)                          |

**Patched record (184 bytes, for patch number 999):**

Bytes `0x00`-`0x3E` are identical to the unpatched record except `counter`
increments (`1` -> `2`); `count` is unchanged. Everything from `0x3F`
onward is replaced with:

| offset | size | field            | value                                                                                                    |
| ------ | ---- | ---------------- | --------------------------------------------------------------------------------------------------------- |
| 0x3F   | 1    | gap byte         | `0x00`                                                                                                    |
| 0x40   | 28   | fixed tag        | constant, not bundle-specific: `EE A8 50 A3 FF 4D E3 D0 AA 23 4C FA FF 07 8A 87 C6 9A B0 BB 00 00 00 00 00 00 00 00` |
| 0x5C   | 4    | count2 (u32 LE)  | 4 (repeat of count at 0x0C)                                                                                |
| 0x60   | 4    | len3 (u32 LE)    | 26 (0x1a)                                                                                                  |
| 0x64   | 26   | str3             | `"9ba626afa44a3aa3.patch_999"`                                                                             |
| 0x7E   | 4    | len4 (u32 LE)    | 33 (0x21)                                                                                                  |
| 0x82   | 33   | str4             | `"9ba626afa44a3aa3.stream.patch_999"`                                                                      |
| 0xA3   | 21   | zero padding     | (out to offset 0xB8 = 184 bytes total)                                                                    |

The only bundle-/patch-specific inputs are the bundle's own hash string
(`"9ba626afa44a3aa3"`, reused verbatim in `str1`-`str4`) and the target
patch number (appears in `str3`/`str4`). The fixed tag, the gap byte, and
the trailing padding are opaque constants observed in a single captured
sample (boot bundle, patch 999), not confirmed universal across other
bundles, patch numbers, or game versions. A different patch number's
digit count changes `len3`/`str3`/`len4`/`str4` and the total record
length; the gap byte and padding sizes are untested beyond
`patchNum=999`.

## Install

```sh
go build -o ./dist/dtpatch ./cmd
```

## Usage

```
dtpatch patch   --dir <bundle dir> [--patch-num N]
dtpatch unpatch --dir <bundle dir> [--patch-num N] [--force]
dtpatch toggle  --dir <bundle dir> [--patch-num N] [--force]
```

`--dir` is required and must point at the directory containing
`bundle_database.data`; there is no install-location auto-detection.
`--patch-num` defaults to `999`. `--force` (unpatch/toggle) skips the
check that the database is currently tagged as patched before restoring
the backup.

## Development

```sh
make test    # go test -v ./...
make build   # CGO_ENABLED=0 go build -o ./dist/dtpatch ./cmd
make ci      # gofmt check, build, vet, race tests, mod tidy check, govulncheck
```

Tests read the real fixture files in `testdata/` directly as golden
data. No synthetic copies are generated. `testdata/bundle_database.data.bak`
is an unpatched capture; `testdata/bundle_database.data` is the same file
already patched at patch number 999, used to verify `dtpatch`'s output
byte-for-byte. `testdata/9ba626afa44a3aa3.patch_999` is a real patch
bundle file kept for reference only; `dtpatch` never touches it.
