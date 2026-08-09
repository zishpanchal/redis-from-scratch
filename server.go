package main

import (
	"fmt"
	"strings"
	"sync"
)

// Server coordinates command execution with durable persistence.
type Server struct {
	handler    *Handler
	aof        *AOF
	mutationMu sync.Mutex
}

func NewServer(handler *Handler, aof *AOF) *Server {
	return &Server{handler: handler, aof: aof}
}

func (s *Server) Execute(request Value) Value {
	if !isValidMutation(request) {
		return s.handler.Handle(request)
	}

	s.mutationMu.Lock()
	defer s.mutationMu.Unlock()

	if err := s.aof.Write(request); err != nil {
		return Value{typ: "error", str: "ERR persistence failure: " + err.Error()}
	}
	return s.handler.Handle(request)
}

func ReplayAOF(aof *AOF, handler *Handler) error {
	return aof.Replay(func(request Value) error {
		result := handler.Handle(request)
		if result.typ == "error" {
			return fmt.Errorf("%s", result.str)
		}
		return nil
	})
}

func isValidMutation(request Value) bool {
	if request.typ != "array" || len(request.array) == 0 || request.array[0].typ != "bulk" {
		return false
	}
	if !allBulkStrings(request.array[1:]) {
		return false
	}

	switch strings.ToUpper(request.array[0].bulk) {
	case "SET":
		return len(request.array) == 3
	case "HSET":
		return len(request.array) == 4
	default:
		return false
	}
}
