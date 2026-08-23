package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/salimnassim/dtpatch"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "patch":
		runPatch(os.Args[2:])
	case "unpatch":
		runUnpatch(os.Args[2:])
	case "toggle":
		runToggle(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: dtpatch <patch|unpatch|toggle> --dir <bundle dir> [--patch-num N] [--force]")
}

func runPatch(args []string) {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory (required)")
	num := fs.Int("patch-num", 999, "patch number to inject")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: dtpatch patch --dir <bundle dir> [--patch-num N]")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	if *dir == "" {
		fs.Usage()
		os.Exit(2)
	}

	if err := dtpatch.Patch(dtpatch.Options{Dir: *dir, PatchNum: *num}); err != nil {
		printFriendlyError(err)
		os.Exit(1)
	}
	fmt.Println("patched", *dir)
}

func runUnpatch(args []string) {
	fs := flag.NewFlagSet("unpatch", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory (required)")
	num := fs.Int("patch-num", 999, "patch number to remove")
	force := fs.Bool("force", false, "unpatch without checking the tag is present")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: dtpatch unpatch --dir <bundle dir> [--patch-num N] [--force]")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	if *dir == "" {
		fs.Usage()
		os.Exit(2)
	}

	if err := dtpatch.Unpatch(dtpatch.Options{Dir: *dir, PatchNum: *num, Force: *force}); err != nil {
		printFriendlyError(err)
		os.Exit(1)
	}
	fmt.Println("unpatched", *dir)
}

func runToggle(args []string) {
	fs := flag.NewFlagSet("toggle", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory (required)")
	num := fs.Int("patch-num", 999, "patch number to toggle")
	force := fs.Bool("force", false, "unpatch without checking the tag is present")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: dtpatch toggle --dir <bundle dir> [--patch-num N] [--force]")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	if *dir == "" {
		fs.Usage()
		os.Exit(2)
	}

	if err := dtpatch.Toggle(dtpatch.Options{Dir: *dir, PatchNum: *num, Force: *force}); err != nil {
		printFriendlyError(err)
		os.Exit(1)
	}
	fmt.Println("toggled", *dir)
}

func printFriendlyError(err error) {
	switch {
	case errors.Is(err, dtpatch.ErrAlreadyPatched):
		fmt.Fprintln(os.Stderr, "already patched; run unpatch first or use toggle")
	case errors.Is(err, dtpatch.ErrNotPatched):
		fmt.Fprintln(os.Stderr, "not patched; nothing to unpatch")
	case errors.Is(err, dtpatch.ErrUnexpectedPatch001):
		fmt.Fprintln(os.Stderr, "unexpected existing patch_001 record; aborting, database not touched")
	case errors.Is(err, dtpatch.ErrNotFound):
		fmt.Fprintln(os.Stderr, "boot bundle record not found; is --dir correct?")
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
