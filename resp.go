package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// Value is one value decoded from the RESP protocol.
type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

// Resp reads RESP values from a buffered stream.
type Resp struct {
	reader *bufio.Reader
}

func NewResp(reader io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(reader)}
}

// Read decodes the next RESP value from the stream.
func (r *Resp) Read() (Value, error) {
	typ, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch typ {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		return Value{}, fmt.Errorf("unsupported RESP type %q", typ)
	}
}

func (r *Resp) readArray() (Value, error) {
	length, _, err := r.readInteger()
	if err != nil {
		return Value{}, fmt.Errorf("read array length: %w", err)
	}
	if length < 0 {
		return Value{}, fmt.Errorf("invalid array length %d", length)
	}

	value := Value{
		typ:   "array",
		array: make([]Value, length),
	}

	for i := range length {
		value.array[i], err = r.Read()
		if err != nil {
			return Value{}, fmt.Errorf("read array element %d: %w", i, err)
		}
	}

	return value, nil
}

func (r *Resp) readBulk() (Value, error) {
	length, _, err := r.readInteger()
	if err != nil {
		return Value{}, fmt.Errorf("read bulk string length: %w", err)
	}
	if length < 0 {
		return Value{}, fmt.Errorf("null bulk strings are not supported yet")
	}

	bulk := make([]byte, length)
	if _, err := io.ReadFull(r.reader, bulk); err != nil {
		return Value{}, fmt.Errorf("read bulk string: %w", err)
	}

	terminator := make([]byte, 2)
	if _, err := io.ReadFull(r.reader, terminator); err != nil {
		return Value{}, fmt.Errorf("read bulk string terminator: %w", err)
	}
	if terminator[0] != '\r' || terminator[1] != '\n' {
		return Value{}, fmt.Errorf("bulk string must end with CRLF")
	}

	return Value{typ: "bulk", bulk: string(bulk)}, nil
}

func (r *Resp) readInteger() (value int, bytesRead int, err error) {
	line, bytesRead, err := r.readLine()
	if err != nil {
		return 0, bytesRead, err
	}

	parsed, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, bytesRead, fmt.Errorf("parse integer %q: %w", line, err)
	}

	return int(parsed), bytesRead, nil
}

func (r *Resp) readLine() (line []byte, bytesRead int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, bytesRead, err
		}

		bytesRead++
		line = append(line, b)

		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			return line[:len(line)-2], bytesRead, nil
		}
	}
}
