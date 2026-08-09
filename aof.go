package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

// AOF stores mutating commands in RESP format for recovery after a restart.
type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open AOF: %w", err)
	}
	return &AOF{file: file}, nil
}

// Write appends and synchronizes one command before it can be acknowledged.
func (aof *AOF) Write(value Value) error {
	data, err := value.Marshal()
	if err != nil {
		return fmt.Errorf("marshal AOF command: %w", err)
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	if _, err := io.Copy(aof.file, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("append AOF command: %w", err)
	}
	if err := aof.file.Sync(); err != nil {
		return fmt.Errorf("sync AOF: %w", err)
	}
	return nil
}

// Replay decodes every stored command in order and passes it to apply.
func (aof *AOF) Replay(apply func(Value) error) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	if _, err := aof.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek to start of AOF: %w", err)
	}

	reader := NewResp(aof.file)
	for {
		value, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("decode AOF: %w", err)
		}
		if err := apply(value); err != nil {
			return fmt.Errorf("apply AOF command: %w", err)
		}
	}
}

func (aof *AOF) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	if err := aof.file.Sync(); err != nil {
		_ = aof.file.Close()
		return fmt.Errorf("sync AOF before close: %w", err)
	}
	return aof.file.Close()
}
