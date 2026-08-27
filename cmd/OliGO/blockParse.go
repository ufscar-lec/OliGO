package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"lec.ufscar.br/OliGO/bio"
)

func runBlockParse(args []string) {
	fs := flag.NewFlagSet("blockParse", flag.ExitOnError)

	var targetFile string
	fs.StringVar(&targetFile, "target", "", "Target FASTA for probe generation.")

	var probeLen int
	fs.IntVar(&probeLen, "len", 40, "Desired probe length (bp); Default is 40.")

	var stepSize int
	fs.IntVar(&stepSize, "step", 1, "Step size for probe generation (bp); Default is 1.")

	var removeMasked bool
	fs.BoolVar(&removeMasked, "removeMasked", false, "Removes probes generated within masked regions.")

	var minProbeGC float64
	fs.Float64Var(&minProbeGC, "minGC", 20, "Minimum GC content per probe (%); Default is 20.")
	var maxProbeGC float64
	fs.Float64Var(&maxProbeGC, "maxGC", 80, "Maximum GC content per probe (%); Default is 80.")

	var maxRepeat int
	fs.IntVar(&maxRepeat, "maxRepeat", 4, "Maximum amount of repeated bases (bp); Default is 4.")

	var strandConc float64
	fs.Float64Var(&strandConc, "st", 25e-9, "Concentration of the strand (mol/L); Default is 25e-9.")
	var naConc float64
	fs.Float64Var(&naConc, "na", 0.39, "Concentration of Na+ (mol/L); Default is 0.39.")
	var formamidePerc float64
	fs.Float64Var(&formamidePerc, "form", 70, "Concentration of formamide (%); Default is 70.")

	var minProbeTm float64
	fs.Float64Var(&minProbeTm, "minTm", 306, "Minimum melting temperature per probe (K); Default is 306.")
	var maxProbeTm float64
	fs.Float64Var(&maxProbeTm, "maxTm", 316, "Maximum melting temperature per probe (K); Default is 316.")

	var outPath string
	fs.StringVar(&outPath, "out", "probes.fasta", `Output FASTA file path; Default is "probes.fasta".`)

	fs.Parse(args)

	file, err := os.Open(targetFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader, err := bio.NewFASTAReader(file)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	for {
		rec, err := reader.Next()
		if errors.Is(err, bio.FASTAErrEmptyInput) {
			log.Fatalf("pipeline stage requires at least one FASTA record, got empty input")
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("reading first sequence: %v", err)
		}

		probes, err := bio.GetProbeCandidates(*rec, probeLen, stepSize, removeMasked, minProbeGC, maxProbeGC, maxRepeat, strandConc, naConc, formamidePerc, minProbeTm, maxProbeTm)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range probes {
			fmt.Fprintf(writer, ">%s\n%s\n", p.Header, string(p.Sequence))
		}
	}
}
