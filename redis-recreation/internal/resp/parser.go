package resp

import (
	"bytes"
	"errors"
	"strconv"
)

type Parser struct {
	buf []byte
	pos int
}

const CRLF_SEPERATOR = "\r\n"

var (
	ErrMalFormedBytes = errors.New("malformed buffer")
	ErrEmptyBuffer    = errors.New("empty buffer")
	ErrIncomplete     = errors.New("icomplelete buffer content")
)

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(buf []byte) ([]string, error) {
	if len(buf) == 0 {
		return nil, ErrEmptyBuffer
	}

	p.buf = buf
	p.pos = 0

	return p.parseArray()
}

func (p *Parser) getLineContentLen(currLine []byte) (int, error) {
	if len(currLine) < 2 {
		return -1, ErrMalFormedBytes
	}

	intRep, err := strconv.Atoi(string(currLine[1:]))
	if err != nil {
		return -1, ErrMalFormedBytes
	}
	return intRep, nil
}

func (p *Parser) parseArray() ([]string, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	if line[0] != '*' {
		return nil, ErrMalFormedBytes
	}

	arrSize, err := p.getLineContentLen(line)
	if err != nil {
		return nil, ErrMalFormedBytes
	}
	res := make([]string, 0, arrSize)

	for range arrSize {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}
		res = append(res, str)
	}

	return res, nil
}

func (p *Parser) parseBulkString() (string, error) {
	line, err := p.readLine()
	if err != nil {
		return "", err
	}

	if line[0] != '$' {
		return "", ErrMalFormedBytes
	}

	actualStringLen, err := p.getLineContentLen(line)
	if err != nil {
		return "", err
	}

	line, err = p.readLine()
	if err != nil {
		return "", err
	}

	if len(line) != actualStringLen {
		return "", ErrMalFormedBytes
	}

	return string(line), nil
}

func (p *Parser) readLine() ([]byte, error) {
	idx := bytes.Index(p.buf[p.pos:], []byte(CRLF_SEPERATOR))
	if idx == -1 {
		return []byte{}, ErrIncomplete
	}

	absIdx := idx + p.pos
	line := p.buf[p.pos:absIdx]

	p.pos = absIdx + len(CRLF_SEPERATOR)

	return line, nil
}
