package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type state struct {
	mu     sync.RWMutex
	values []float64
}

func NewState() *state {
	return &state{
		mu:     sync.RWMutex{},
		values: []float64{},
	}
}

func (s *state) add(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, v)
}

func (s *state) get() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values
}

func main() {
	n := maelstrom.NewNode()
	appState := NewState()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"].(float64)
		appState.add(message)

		if err := broadcast(n, msg.Src, body); err != nil {
			return err
		}

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		res := map[string]any{
			"type":     "read_ok",
			"messages": appState.get(),
		}

		return n.Reply(msg, res)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		res := map[string]any{
			"type": "topology_ok",
		}

		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

func broadcast(n *maelstrom.Node, src string, body map[string]any) error {
	for _, dest := range n.NodeIDs() {
		if dest == src || dest == n.ID() {
			continue
		}

		go func(dest string) {
			if err := n.Send(dest, body); err != nil {
				panic(err)
			}
		}(dest)
	}
	return nil
}
