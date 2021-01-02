package rvnodegen

import (
	"context"
	"fmt"
	"time"
)

// WebsocketWriter is an interface for writing to a web socket.
type WebsocketWriter interface {
	// Write writes to a web socket.
	Write(r WebsocketResponse) error
}

// CommandHandler is a handler for a command.
type CommandHandler interface {
	// Name is the name of the command.
	Name() string
	// Run runs the action that accepts a command.
	Run(ctx context.Context, w WebsocketWriter, c Command) error
}

// CommandsFactory is a factory for generating a list of command handlers.
func CommandsFactory(lister Lister) []CommandHandler {
	return []CommandHandler{
		NewWorkloadsCommand(lister),
	}
}

// WorkloadsCommand is a workloads command.
type WorkloadsCommand struct {
	lister Lister
}

var _ CommandHandler = &WorkloadsCommand{}

// NewWorkloadsCommand creates an instance of WorkloadsCommand.
func NewWorkloadsCommand(lister Lister) *WorkloadsCommand {
	w := &WorkloadsCommand{
		lister: lister,
	}
	return w
}

// Name returns the name of the handler.
func (wc *WorkloadsCommand) Name() string {
	return "workloads"
}

// Run runs the handler.
func (wc *WorkloadsCommand) Run(ctx context.Context, w WebsocketWriter, c Command) error {
	timer := time.NewTimer(0)
	done := false

	namespace, ok := c.Payload["namespace"].(string)
	if !ok {
		return fmt.Errorf("payload does not have a namespace")
	}

	for !done {
		select {
		case <-ctx.Done():
			done = true
			break
		case <-timer.C:
			nb := NewNodeBuilder(wc.lister)
			nodes, err := nb.Build(namespace)
			if err != nil {
				return fmt.Errorf("build nodes: %w", err)
			}

			payload := map[string]interface{}{
				"type": "nodes",
				"data": map[string]interface{}{
					"nodes": nodes,
				},
			}
			resp := c.CreateResponse(payload)

			if err := w.Write(resp); err != nil {
				return err
			}

			timer.Reset(1 * time.Second)
		}
	}

	return nil
}
