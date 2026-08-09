package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadArrayOfBulkStrings(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$5\r\nadmin\r\n$5\r\nahmed\r\n"

	value, err := NewResp(strings.NewReader(input)).Read()
	if err != nil {
		t.Fatalf("Read() returned an error: %v", err)
	}
	if value.typ != "array" {
		t.Fatalf("type = %q, want array", value.typ)
	}

	want := []string{"SET", "admin", "ahmed"}
	if len(value.array) != len(want) {
		t.Fatalf("array length = %d, want %d", len(value.array), len(want))
	}
	for i, expected := range want {
		if value.array[i].typ != "bulk" || value.array[i].bulk != expected {
			t.Errorf("array[%d] = %#v, want bulk string %q", i, value.array[i], expected)
		}
	}
}

func TestReadRejectsTruncatedBulkString(t *testing.T) {
	_, err := NewResp(strings.NewReader("$5\r\nabc")).Read()
	if err == nil {
		t.Fatal("Read() returned nil error for a truncated bulk string")
	}
}

func TestReadRejectsUnknownType(t *testing.T) {
	_, err := NewResp(strings.NewReader("!wat\r\n")).Read()
	if err == nil {
		t.Fatal("Read() returned nil error for an unknown RESP type")
	}
}

func TestValueMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "simple string", value: Value{typ: "string", str: "OK"}, want: "+OK\r\n"},
		{name: "error", value: Value{typ: "error", str: "ERR unknown command"}, want: "-ERR unknown command\r\n"},
		{name: "integer", value: Value{typ: "integer", num: 42}, want: ":42\r\n"},
		{name: "bulk string", value: Value{typ: "bulk", bulk: "hello"}, want: "$5\r\nhello\r\n"},
		{name: "null", value: Value{typ: "null"}, want: "$-1\r\n"},
		{
			name: "array",
			value: Value{typ: "array", array: []Value{
				{typ: "bulk", bulk: "hello"},
				{typ: "integer", num: 7},
			}},
			want: "*2\r\n$5\r\nhello\r\n:7\r\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.value.Marshal()
			if err != nil {
				t.Fatalf("Marshal() returned an error: %v", err)
			}
			if string(got) != test.want {
				t.Errorf("Marshal() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWriterWritesMarshaledValue(t *testing.T) {
	var destination bytes.Buffer

	err := NewWriter(&destination).Write(Value{typ: "bulk", bulk: "hello"})
	if err != nil {
		t.Fatalf("Write() returned an error: %v", err)
	}
	if got, want := destination.String(), "$5\r\nhello\r\n"; got != want {
		t.Errorf("written bytes = %q, want %q", got, want)
	}
}
