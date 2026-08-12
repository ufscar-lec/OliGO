package bio

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const maxSAMLineSize = 256 * 1024 * 1024

var SAMErrEmptyInput = errors.New("sam: empty input")

type SAMRecord struct {
	QName string
	Flag  uint16
	RName string
	Pos   int
	MAPQ  uint8
	Cigar string
	RNext string
	PNext int
	TLen  int
	Seq   []byte
	Qual  []byte
	Tags  []byte
	AS    int
	HasAS bool
}

type SAMReader struct {
	scanner    *bufio.Scanner
	pending    string
	hasPending bool
	lineNo     int
	sawAnyLine bool
}

func NewSAMReader(r io.Reader) (*SAMReader, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxSAMLineSize)
	reader := &SAMReader{scanner: sc}

	for sc.Scan() {
		reader.lineNo++
		line := sc.Text()
		if reader.lineNo == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		if line == "" {
			continue
		}
		reader.sawAnyLine = true

		if line[0] != '@' {
			reader.pending, reader.hasPending = line, true
			break
		}
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("sam: line %d exceeds max buffer size (%d bytes) - file may be binary or corrupted", reader.lineNo, maxSAMLineSize)
		}
		return nil, err
	}
	if !reader.sawAnyLine {
		return nil, fmt.Errorf("%w: input contained no data", SAMErrEmptyInput)
	}
	return reader, nil
}

func (reader *SAMReader) Next() (*SAMRecord, error) {
	var line []byte
	if reader.hasPending {
		line, reader.hasPending = []byte(reader.pending), false
	} else {
		for {
			if !reader.scanner.Scan() {
				if err := reader.scanner.Err(); err != nil {
					if errors.Is(err, bufio.ErrTooLong) {
						return nil, fmt.Errorf("sam: line %d exceeds max buffer size (%d bytes) - file may be binary or corrupted", reader.lineNo, maxSAMLineSize)
					}
					return nil, err
				}
				return nil, io.EOF
			}
			reader.lineNo++
			if candidate := bytes.TrimSpace(reader.scanner.Bytes()); len(candidate) > 0 {
				line = candidate
				break
			}
		}
	}
	return reader.parseRecord(line)
}

func (reader *SAMReader) parseRecord(line []byte) (*SAMRecord, error) {
	fields := bytes.SplitN(line, []byte("\t"), 12)

	if len(fields) < 11 {
		return nil, fmt.Errorf("sam: line %d has %d fields, need at least 11; input does not look like SAM", reader.lineNo, len(fields))
	}

	flag, err := strconv.ParseUint(string(fields[1]), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("sam: line %d has invalid FLAG %q: %w", reader.lineNo, fields[1], err)
	}

	pos, err := strconv.Atoi(string(fields[3]))
	if err != nil {
		return nil, fmt.Errorf("sam: line %d has invalid POS %q: %w", reader.lineNo, fields[3], err)
	}

	mapq, err := strconv.ParseUint(string(fields[4]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("sam: line %d has invalid MAPQ %q: %w", reader.lineNo, fields[4], err)
	}

	pnext, err := strconv.Atoi(string(fields[7]))
	if err != nil {
		return nil, fmt.Errorf("sam: line %d has invalid PNEXT %q: %w", reader.lineNo, fields[7], err)
	}

	tlen, err := strconv.Atoi(string(fields[8]))
	if err != nil {
		return nil, fmt.Errorf("sam: line %d has invalid TLEN %q: %w", reader.lineNo, fields[8], err)
	}

	rec := &SAMRecord{
		QName: string(fields[0]),
		Flag:  uint16(flag),
		RName: string(fields[2]),
		Pos:   pos,
		MAPQ:  uint8(mapq),
		Cigar: string(fields[5]),
		RNext: string(fields[6]),
		PNext: pnext,
		TLen:  tlen,
		Seq:   append([]byte(nil), fields[9]...),
		Qual:  append([]byte(nil), fields[10]...),
	}

	if len(fields) == 12 {
		rec.Tags = append([]byte(nil), fields[11]...)
		for tag := range bytes.SplitSeq(fields[11], []byte("\t")) {
			if bytes.HasPrefix(tag, []byte("AS:i:")) {
				val, convErr := strconv.Atoi(string(tag[len("AS:i:"):]))
				if convErr == nil {
					rec.AS = val
					rec.HasAS = true
				}
				break
			}
		}
	}

	return rec, nil
}
