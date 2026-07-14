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
	flag.StringVar(&targetFile, "target", "", "Target SAM for probe filtering.")

	var minMAPQ uint
	flag.UintVar(&minMAPQ, "minMAPQ", 30, "Minimum required MAPQ; Default is 30.")

	var uniqueOnly bool
	flag.BoolVar(&uniqueOnly, "unique", false, "Removes non-unique probes.")

	var outPath string
	flag.StringVar(&outPath, "out", "filtered_probes.fasta", "Output FASTA file path; Default is \"filtered_probes.fasta\".")

	flag.Parse()

	file, err := os.Open(targetFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader, err := bio.NewSAMReader(file)
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

	if uniqueOnly {
		counts := make(map[string]int)

		for {
			rec, err := reader.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Fatalf("counting pass: %v", err)
			}
			counts[rec.QName]++
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Fatal(err)
		}

		reader, err = bio.NewSAMReader(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		rec, err := reader.Next()
		if errors.Is(err, bio.SAMErrEmptyInput) {
			log.Fatalf("pipeline stage requires at least one SAM record, got empty input")
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("reading first sequence: %v", err)
		}

		if rec.Flag != 0 || rec.MAPQ < uint8(minMAPQ) {
			continue
		}

		fmt.Fprintf(writer, ">%s\n%s\n", rec.QName, string(rec.Seq))
	}
}
