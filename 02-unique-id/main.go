package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// generate a random number between 0 and 999999999999999999
		randomNum, err := rand.Int(rand.Reader, big.NewInt(1000000000000000000))
		if err != nil {
			fmt.Println("error generating random number:", err)
			return err
		}

		// Update the message type.
		body["type"] = "generate_ok"
		body["id"] = randomNum

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
