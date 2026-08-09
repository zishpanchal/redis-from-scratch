package main

import (
	"fmt"
	"strings"
	"sync"
)

// Handler owns the in-memory data and dispatches Redis commands.
type Handler struct {
	stringsMu sync.RWMutex
	strings   map[string]string
}

func NewHandler() *Handler {
	return &Handler{strings: make(map[string]string)}
}

func (h *Handler) Handle(request Value) Value {
	if request.typ != "array" || len(request.array) == 0 {
		return commandError("expected a non-empty command array")
	}
	if request.array[0].typ != "bulk" {
		return commandError("command name must be a bulk string")
	}

	command := strings.ToUpper(request.array[0].bulk)
	args := request.array[1:]
	if !allBulkStrings(args) {
		return commandError("command arguments must be bulk strings")
	}

	switch command {
	case "PING":
		return h.ping(args)
	case "SET":
		return h.set(args)
	case "GET":
		return h.get(args)
	default:
		return Value{typ: "error", str: fmt.Sprintf("ERR unknown command '%s'", strings.ToLower(command))}
	}
}

func (h *Handler) ping(args []Value) Value {
	switch len(args) {
	case 0:
		return Value{typ: "string", str: "PONG"}
	case 1:
		return Value{typ: "bulk", bulk: args[0].bulk}
	default:
		return wrongNumberOfArguments("ping")
	}
}

func (h *Handler) set(args []Value) Value {
	if len(args) != 2 {
		return wrongNumberOfArguments("set")
	}

	h.stringsMu.Lock()
	h.strings[args[0].bulk] = args[1].bulk
	h.stringsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func (h *Handler) get(args []Value) Value {
	if len(args) != 1 {
		return wrongNumberOfArguments("get")
	}

	h.stringsMu.RLock()
	value, exists := h.strings[args[0].bulk]
	h.stringsMu.RUnlock()

	if !exists {
		return Value{typ: "null"}
	}
	return Value{typ: "bulk", bulk: value}
}

func allBulkStrings(values []Value) bool {
	for _, value := range values {
		if value.typ != "bulk" {
			return false
		}
	}
	return true
}

func commandError(message string) Value {
	return Value{typ: "error", str: "ERR " + message}
}

func wrongNumberOfArguments(command string) Value {
	return Value{typ: "error", str: fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)}
}
