package main

import (
	"log"
	"os"

	"github.com/andrewwtpowell/dfs/pkg/client"
)

var (
	server = os.Getenv("DFS_SERVER_ADDR")
)

func main() {

    dfsClient, err := client.Init(server)
    if err != nil {
        log.Fatalf("Failure initializing client: %v", err)
    }

    defer dfsClient.Shutdown()


	if len(os.Args) < 2 {
		log.Fatal("Expected subcommand. Subcommand options are list, store, stat, delete, or fetch.")
	}

	switch os.Args[1] {
	case "mount":
		mountPath := os.Args[2]
		if mountPath == "" {
			mountPath = "mnt/"
		}
		dfsClient.Mount(&mountPath)
	case "list":
		dfsClient.List()
	case "store":
		filename := os.Args[2]
		if filename == "" {
			log.Fatal("No file name provided for store RPC")
		}
		dfsClient.Store(&filename)
	case "stat":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for stat RPC")
		}
		dfsClient.Stat(&filename)
	case "delete":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for delete RPC")
		}
		dfsClient.DeleteFile(&filename)
	case "fetch":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for fetch RPC")
		}
		dfsClient.Fetch(&filename)
	default:
		log.Fatal("Invalid subcommand. Expected list, store, stat, delete, or fetch.")
	}
}
