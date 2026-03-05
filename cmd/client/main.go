package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/liambrem/dist-kv-store-raft/kv"
)

func main() {
	port := flag.Int("port", 9000, "HTTP API port")
	host := flag.String("host", "localhost", "Server host")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/client/main.go put <key> <value>")
		fmt.Println("  go run cmd/client/main.go get <key>")
		fmt.Println("  go run cmd/client/main.go keys")
		os.Exit(1)
	}

	client := kv.NewClient(*host, *port)

	command := args[0]

	switch command {
	case "put":
		if len(args) != 3 {
			log.Fatal("Usage: put <key> <value>")
		}
		key := args[1]
		value := args[2]

		if err := client.Put(key, value); err != nil {
			log.Fatalf("PUT failed: %v", err)
		}
		fmt.Printf("OK: Set %s = %s\n", key, value)

	case "get":
		if len(args) != 2 {
			log.Fatal("Usage: get <key>")
		}
		key := args[1]

		value, err := client.Get(key)
		if err != nil {
			log.Fatalf("GET failed: %v", err)
		}
		fmt.Printf("%s = %s\n", key, value)

	case "keys":
		keys, err := client.GetAllKeys()
		if err != nil {
			log.Fatalf("KEYS failed: %v", err)
		}
		fmt.Printf("Keys (%d):\n", len(keys))
		for _, k := range keys {
			fmt.Printf("  - %s\n", k)
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
