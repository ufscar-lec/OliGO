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

func main() {
	var targetFile string
	flag.StringVar(&targetFile, "target", "", "Target FASTA for probe generation.")

	var probeLen int
	flag.IntVar(&probeLen, "len", 40, "Desired probe length (bp); Default is 40.")

	var stepSize int
	flag.IntVar(&stepSize, "step", 1, "Step size for probe generation (bp); Default is 1.")

	var removeMasked bool
	flag.BoolVar(&removeMasked, "removeMasked", false, "Removes probes generated within masked regions.")

	var minProbeGC float64
	flag.Float64Var(&minProbeGC, "minGC", 20, "Minimum GC content per probe (%); Default is 20.")
	var maxProbeGC float64
	flag.Float64Var(&maxProbeGC, "maxGC", 80, "Maximum GC content per probe (%); Default is 80.")

	var maxRepeat int
	flag.IntVar(&maxRepeat, "maxRepeat", 4, "Maximum amount of repeated bases (bp); Default is 4.")

	var strandConc float64
	flag.Float64Var(&strandConc, "st", 25e-9, "Concentration of the strand (mol/L); Default is 25e-9.")
	var naConc float64
	flag.Float64Var(&naConc, "na", 0.39, "Concentration of Na+ (mol/L); Default is 0.39.")

	var minProbeTm float64
	flag.Float64Var(&minProbeTm, "minTm", 350, "Minimum melting temperature per probe (K); Default is 350.")
	var maxProbeTm float64
	flag.Float64Var(&maxProbeTm, "maxTm", 354, "Maximum melting temperature per probe (K); Default is 354.")

	var outPath string
	flag.StringVar(&outPath, "out", "probes.fasta", "Output FASTA file path; Default is \"probes.fasta\".")

	flag.Parse()

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

		probes, err := bio.GetProbeCandidates(*rec, probeLen, stepSize, removeMasked, minProbeGC, maxProbeGC, maxRepeat, strandConc, naConc, minProbeTm, maxProbeTm)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range probes {
			fmt.Fprintf(writer, ">%s\n%s\n", p.Header, string(p.Sequence))
		}
	}
}
