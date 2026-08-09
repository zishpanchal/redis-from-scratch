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
}
