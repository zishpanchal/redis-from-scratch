package main

import "testing"

func TestHandlerPing(t *testing.T) {
	handler := NewHandler()

	assertValue(t, handler.Handle(command("ping")), Value{typ: "string", str: "PONG"})
	assertValue(t, handler.Handle(command("PING", "hello")), Value{typ: "bulk", bulk: "hello"})
	assertValue(t, handler.Handle(command("PING", "one", "two")), wrongNumberOfArguments("ping"))
}

func TestHandlerSetAndGet(t *testing.T) {
	handler := NewHandler()

	assertValue(t, handler.Handle(command("SET", "language", "Go")), Value{typ: "string", str: "OK"})
	assertValue(t, handler.Handle(command("GET", "language")), Value{typ: "bulk", bulk: "Go"})
	assertValue(t, handler.Handle(command("GET", "missing")), Value{typ: "null"})
}

func TestHandlerHashCommands(t *testing.T) {
	handler := NewHandler()

	assertValue(t, handler.Handle(command("HSET", "users", "u2", "Grace")), Value{typ: "integer", num: 1})
	assertValue(t, handler.Handle(command("HSET", "users", "u1", "Ada")), Value{typ: "integer", num: 1})
	assertValue(t, handler.Handle(command("HSET", "users", "u1", "Ada Lovelace")), Value{typ: "integer", num: 0})
	assertValue(t, handler.Handle(command("HGET", "users", "u1")), Value{typ: "bulk", bulk: "Ada Lovelace"})
	assertValue(t, handler.Handle(command("HGET", "users", "missing")), Value{typ: "null"})

	got := handler.Handle(command("HGETALL", "users"))
	want := Value{typ: "array", array: []Value{
		{typ: "bulk", bulk: "u1"},
		{typ: "bulk", bulk: "Ada Lovelace"},
		{typ: "bulk", bulk: "u2"},
		{typ: "bulk", bulk: "Grace"},
	}}
	assertValue(t, got, want)
}

func TestHandlerRejectsInvalidCommands(t *testing.T) {
	handler := NewHandler()

	assertValue(t, handler.Handle(Value{typ: "bulk", bulk: "PING"}), commandError("expected a non-empty command array"))
	assertValue(t, handler.Handle(command("NOPE")), Value{typ: "error", str: "ERR unknown command 'nope'"})
	assertValue(t, handler.Handle(command("SET", "only-a-key")), wrongNumberOfArguments("set"))
}

func command(name string, args ...string) Value {
	values := make([]Value, 0, len(args)+1)
	values = append(values, Value{typ: "bulk", bulk: name})
	for _, arg := range args {
		values = append(values, Value{typ: "bulk", bulk: arg})
	}
	return Value{typ: "array", array: values}
}

func assertValue(t *testing.T, got, want Value) {
	t.Helper()
	if got.typ != want.typ || got.str != want.str || got.num != want.num || got.bulk != want.bulk {
		t.Errorf("value = %#v, want %#v", got, want)
	}
	if len(got.array) != len(want.array) {
		t.Fatalf("array length = %d, want %d", len(got.array), len(want.array))
	}
	for i := range want.array {
		assertValue(t, got.array[i], want.array[i])
	}
}
