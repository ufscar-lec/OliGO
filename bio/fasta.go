package bio

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

const maxFASTALineSize = 64 * 1024 * 1024

var FASTAErrEmptyInput = errors.New("fasta: empty input")

type FASTARecord struct {
	Header   string
	Sequence []byte
}

type FASTAReader struct {
	scanner    *bufio.Scanner
	pending    string
	hasPending bool
	lineNo     int
	sawAnyLine bool
}

func NewFASTAReader(reader io.Reader) (*FASTAReader, error) {
	if reader == nil {
		return nil, errors.New("fasta: reader is nil")
	}
	sc := bufio.NewScanner(reader)
	sc.Buffer(make([]byte, 0, 64*1024), maxFASTALineSize)
	return &FASTAReader{scanner: sc}, nil
}

func (reader *FASTAReader) Next() (*FASTARecord, error) {
	var hed string
	var hasHed bool
	var seq []byte

	if reader.hasPending {
		hed, hasHed = reader.pending, true
		reader.pending, reader.hasPending = "", false
	}

	for reader.scanner.Scan() {
		reader.lineNo++
		line := strings.TrimSpace(reader.scanner.Text())
		if reader.lineNo == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		if line == "" {
			continue
		}
		reader.sawAnyLine = true
		if line[0] == '>' {
			if !hasHed {
				hed, hasHed = line[1:], true
				continue
			}
			reader.pending, reader.hasPending = line[1:], true
			return &FASTARecord{Header: hed, Sequence: seq}, nil
		}
		if !hasHed {
			return nil, fmt.Errorf("fasta: line %d has sequence data before any header ('>'); input does not look like FASTA", reader.lineNo)
		}
		seq = append(seq, line...)
	}

	if err := reader.scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("fasta: line %d exceeds max buffer size (%d bytes) - file may be binary or corrupted", reader.lineNo, maxFASTALineSize)
		}
		return nil, err
	}

	if !hasHed {
		if !reader.sawAnyLine {
			return nil, fmt.Errorf("%w: input contained no data", FASTAErrEmptyInput)
		}
		return nil, io.EOF
	}

	return &FASTARecord{Header: hed, Sequence: seq}, nil
}

func GetProbeCandidates(rec FASTARecord, length int, step int, removeMasked bool, minGC float64, maxGC float64, maxRepeat int, strandConc float64, naConc float64, formamidePerc float64, minProbeTm float64, maxProbeTm float64) ([]FASTARecord, error) {
	var recordList []FASTARecord
	var saltCorr = 16.6 * math.Log10(naConc)

	var deltaHTable = map[[2]byte]float64{
		{'A', 'A'}: -7.9,
		{'T', 'T'}: -7.9,
		{'A', 'T'}: -7.2,
		{'T', 'A'}: -7.2,
		{'C', 'A'}: -8.5,
		{'T', 'G'}: -8.5, // revcomp of CA
		{'G', 'T'}: -8.4,
		{'A', 'C'}: -8.4, // revcomp of GT
		{'C', 'T'}: -7.8,
		{'A', 'G'}: -7.8, // revcomp of CT
		{'G', 'A'}: -8.2,
		{'T', 'C'}: -8.2, // revcomp of GA
		{'C', 'G'}: -10.6,
		{'G', 'C'}: -9.8,
		{'G', 'G'}: -8.0,
		{'C', 'C'}: -8.0,
	}

	var deltaSTable = map[[2]byte]float64{
		{'A', 'A'}: -22.2,
		{'T', 'T'}: -22.2,
		{'A', 'T'}: -20.4,
		{'T', 'A'}: -21.3,
		{'C', 'A'}: -22.8,
		{'T', 'G'}: -22.8, // revcomp of CA
		{'G', 'T'}: -22.4,
		{'A', 'C'}: -22.4, // revcomp of GT
		{'C', 'T'}: -21.0,
		{'A', 'G'}: -21.0, // revcomp of CT
		{'G', 'A'}: -22.2,
		{'T', 'C'}: -22.2, // revcomp of GA
		{'C', 'G'}: -27.2,
		{'G', 'C'}: -24.4,
		{'G', 'G'}: -19.9,
		{'C', 'C'}: -19.9,
	}

	if step <= 0 {
		return nil, fmt.Errorf("step must be greater than 0, got %d", step)
	}

	if maxRepeat <= 0 {
		return nil, fmt.Errorf("maxRepeat must be greater than 0, got %d", maxRepeat)
	}

	for index := 0; index <= len(rec.Sequence)-length; index += step {
		probe := FASTARecord{
			Header:   "candidate_" + strconv.Itoa(index),
			Sequence: rec.Sequence[index : index+length],
		}

		isValid := bytes.IndexFunc(probe.Sequence, func(r rune) bool {
			if removeMasked {
				return r != 'A' && r != 'C' && r != 'T' && r != 'G'
			}
			return r != 'A' && r != 'C' && r != 'T' && r != 'G' && r != 'a' && r != 'c' && r != 't' && r != 'g'
		}) == -1

		if !isValid {
			continue
		}

		hasHomopolymerRun := false
		runLength := 1
		countGC := 0

		for i, base := range probe.Sequence {
			if base == 'G' || base == 'C' {
				countGC++
			}

			if i > 0 {
				if base == probe.Sequence[i-1] {
					runLength++
					if runLength > maxRepeat {
						hasHomopolymerRun = true
						break
					}
				} else {
					runLength = 1
				}
			}
		}

		if hasHomopolymerRun {
			continue
		}

		percentageGC := (float64(countGC) / float64(len(probe.Sequence))) * 100

		if percentageGC < minGC || percentageGC > maxGC {
			continue
		}

		deltaH := 0.0
		deltaS := 0.0

		probe.Sequence = bytes.ToUpper(probe.Sequence)

		for i := 0; i < len(probe.Sequence)-1; i++ {
			key := [2]byte{probe.Sequence[i], probe.Sequence[i+1]}

			hVal, hOk := deltaHTable[key]
			if !hOk {
				return nil, fmt.Errorf("no deltaH NN parameter for dinucleotide %s", string(key[:]))
			}
			deltaH += hVal

			sVal, sOk := deltaSTable[key]
			if !sOk {
				return nil, fmt.Errorf("no deltaS NN parameter for dinucleotide %s", string(key[:]))
			}
			deltaS += sVal
		}

		if probe.Sequence[0] == 'A' || probe.Sequence[0] == 'T' {
			deltaH += 2.3
			deltaS += 4.1
		} else {
			deltaH += 0.1
			deltaS += -2.8
		}

		if probe.Sequence[len(probe.Sequence)-1] == 'A' || probe.Sequence[len(probe.Sequence)-1] == 'T' {
			deltaH += 2.3
			deltaS += 4.1
		} else {
			deltaH += 0.1
			deltaS += -2.8
		}

		probeTm := (deltaH*1000)/(deltaS+1.987*math.Log(strandConc/4)) +
			saltCorr - 0.65*formamidePerc

		if probeTm > maxProbeTm || probeTm < minProbeTm {
			continue
		}

		recordList = append(recordList, probe)
	}

	return recordList, nil
}
