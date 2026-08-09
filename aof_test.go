package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAOFRestoresMutations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "database.aof")
	aof, err := NewAOF(path)
	if err != nil {
		t.Fatalf("NewAOF() returned an error: %v", err)
	}

	server := NewServer(NewHandler(), aof)
	assertValue(t, server.Execute(command("SET", "language", "Go")), Value{typ: "string", str: "OK"})
	assertValue(t, server.Execute(command("HSET", "users", "u1", "Ada")), Value{typ: "integer", num: 1})
	server.Execute(command("GET", "language"))
	server.Execute(command("PING"))

	if err := aof.Close(); err != nil {
		t.Fatalf("Close() returned an error: %v", err)
	}

	reopened, err := NewAOF(path)
	if err != nil {
		t.Fatalf("reopen AOF: %v", err)
	}
	defer reopened.Close()

	restored := NewHandler()
	if err := ReplayAOF(reopened, restored); err != nil {
		t.Fatalf("ReplayAOF() returned an error: %v", err)
	}

	assertValue(t, restored.Handle(command("GET", "language")), Value{typ: "bulk", bulk: "Go"})
	assertValue(t, restored.Handle(command("HGET", "users", "u1")), Value{typ: "bulk", bulk: "Ada"})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read AOF: %v", err)
	}
	want := "*3\r\n$3\r\nSET\r\n$8\r\nlanguage\r\n$2\r\nGo\r\n" +
		"*4\r\n$4\r\nHSET\r\n$5\r\nusers\r\n$2\r\nu1\r\n$3\r\nAda\r\n"
	if string(data) != want {
		t.Errorf("AOF contents = %q, want %q", data, want)
	}
}

func TestAOFRejectsCorruptData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "database.aof")
	if err := os.WriteFile(path, []byte("*1\r\n$5\r\nabc"), 0o644); err != nil {
		t.Fatalf("write corrupt AOF: %v", err)
	}

	aof, err := NewAOF(path)
	if err != nil {
		t.Fatalf("NewAOF() returned an error: %v", err)
	}
	defer aof.Close()

	if err := ReplayAOF(aof, NewHandler()); err == nil {
		t.Fatal("ReplayAOF() returned nil error for corrupt data")
	}
}
