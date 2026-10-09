package raft

import (
	"encoding/json"
	"errors"
)

// Command represents a deterministic state-machine operation.
type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// EncodeCommand serializes a command for storage in the Raft log.
func EncodeCommand(command Command) (string, error) {
	if command.Key == "" {
		return "", errors.New("command key cannot be empty")
	}

	switch command.Op {
	case "put":
	case "delete":
		if command.Value != "" {
			return "", errors.New("delete command cannot contain a value")
		}
	default:
		return "", errors.New("unsupported command operation")
	}

	data, err := json.Marshal(command)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeCommand validates and decodes a Raft log command.
func DecodeCommand(raw string) (Command, error) {
	var command Command
	if err := json.Unmarshal([]byte(raw), &command); err != nil {
		return Command{}, err
	}

	if command.Key == "" {
		return Command{}, errors.New("command key cannot be empty")
	}

	switch command.Op {
	case "put":
	case "delete":
		if command.Value != "" {
			return Command{}, errors.New("delete command cannot contain a value")
		}
	default:
		return Command{}, errors.New("unsupported command operation")
	}

	return command, nil
}
