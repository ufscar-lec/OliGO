package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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

	var singleChr string
	flag.StringVar(&singleChr, "singleChr", "", "If provided, remove the probes that are not mapped to said chromosome.")

	var minStartPos int
	flag.IntVar(&minStartPos, "minStartPos", -1, "If provided, remove the probes that start before said position (exclusively).")

	var maxEndPos int
	flag.IntVar(&maxEndPos, "maxEndPos", -1, "If provided, remove the probes that end after said position (exclusively).")

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

	hits := make(map[string]*struct {
		seq []byte
		pos []string
	})
	var order []string

	for {
		rec, err := reader.Next()
		if errors.Is(err, bio.SAMErrEmptyInput) {
			log.Fatalf("pipeline stage requires at least one SAM record, got empty input")
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("reading sequence: %v", err)
		}

		if rec.Flag != 0 || rec.MAPQ < uint8(minMAPQ) {
			continue
		}

		id := rec.QName
		if i := strings.IndexByte(id, '|'); i != -1 {
			id = id[:i]
		}

		refSpan := 0
		numBuf := 0
		hasNum := false
		for _, c := range rec.Cigar {
			if c >= '0' && c <= '9' {
				numBuf = numBuf*10 + int(c-'0')
				hasNum = true
				continue
			}
			if hasNum {
				switch c {
				case 'M', 'D', 'N', '=', 'X':
					refSpan += numBuf
				}
			}
			numBuf = 0
			hasNum = false
		}
		if refSpan == 0 {
			refSpan = len(rec.Seq)
		}

		startPos := rec.Pos
		endPos := rec.Pos + refSpan - 1
		posStr := fmt.Sprintf("%s:%d-%d", rec.RName, startPos, endPos)

		if singleChr != "" && rec.RName != singleChr {
			continue
		}

		if minStartPos != -1 && startPos < minStartPos {
			continue
		}

		if maxEndPos != -1 && endPos > maxEndPos {
			continue
		}

		h, ok := hits[id]
		if !ok {
			h = &struct {
				seq []byte
				pos []string
			}{seq: rec.Seq}
			hits[id] = h
			order = append(order, id)
		}
		h.pos = append(h.pos, posStr)
	}

	for _, id := range order {
		h := hits[id]
		if uniqueOnly && len(h.pos) != 1 {
			continue
		}
		fmt.Fprintf(writer, ">%s|%s\n%s\n", id, strings.Join(h.pos, ";"), string(h.seq))
	}
}
