package main

import (
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
