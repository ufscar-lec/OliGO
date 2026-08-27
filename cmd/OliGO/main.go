package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: OliGO <command> [arguments]")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  blockParse    generate probe candidates from a FASTA file")
		fmt.Fprintln(os.Stderr, "  filterProbes  filter mapped probes from a SAM file")
		os.Exit(1)
	}

	cmd, args := os.Args[1], os.Args[2:]

	switch cmd {
	case "blockParse":
		runBlockParse(args)
	case "filterProbes":
		runFilterProbes(args)
	default:
		fmt.Fprintf(os.Stderr, "OliGO: unknown command %q\n", cmd)
		os.Exit(1)
	}
}
