package rvnodegen

import (
	"context"
	"fmt"
	"time"
)

type WebsocketWriter interface {
	Write(r WebsocketResponse) error
}

type CommandHandler interface {
	Name() string
	Run(ctx context.Context, w WebsocketWriter, c Command) error
}

func CommandsFactory(lister Lister) []CommandHandler {
	return []CommandHandler{
		NewWorkloadsCommand(lister),
	}
}

type WorkloadsCommand struct {
	lister Lister
}

var _ CommandHandler = &WorkloadsCommand{}

func NewWorkloadsCommand(lister Lister) *WorkloadsCommand {
	w := &WorkloadsCommand{
		lister: lister,
	}
	return w
}

func (wc *WorkloadsCommand) Name() string {
	return "workloads"
}

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
