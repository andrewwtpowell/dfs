package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"

	pb "github.com/andrewwtpowell/dfs/api/dfs_api"
	"github.com/andrewwtpowell/dfs/pkg/client"
)

var (
	server = os.Getenv("DFS_SERVER_ADDR")
)

func main() {

	if server == "" {
		server = "localhost:50051"
		log.Printf("DFS_SERVER_ADDR env var not set. Using default: %s", server)
	}

	name, err := os.Hostname()
	if err != nil {
		log.Fatalf("os.Hostname: %s", err)
	}
	pid := os.Getpid()
	id := name + fmt.Sprint(pid)
	log.Printf("starting client %s", id)

	// Connect to server
	log.Printf("connecting to server at %s", server)
	conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("unable to connect: %s", err)
	}
	defer conn.Close()
	dfsClient := pb.NewDFSClient(conn)

	if len(os.Args) < 2 {
		log.Fatal("Expected subcommand. Subcommand options are list, store, stat, delete, or fetch.")
	}

	switch os.Args[1] {
	case "mount":
		mountPath := os.Args[2]
		if mountPath == "" {
			mountPath = "mnt/"
		}
		client.Mount(dfsClient, &mountPath)
	case "list":
		client.List(dfsClient)
	case "store":
		filename := os.Args[2]
		if filename == "" {
			log.Fatal("No file name provided for store RPC")
		}
		client.Store(dfsClient, &filename)
	case "stat":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for stat RPC")
		}
		client.Stat(dfsClient, &filename)
	case "delete":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for delete RPC")
		}
		client.DeleteFile(dfsClient, &filename)
	case "fetch":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for fetch RPC")
		}
		client.Fetch(dfsClient, &filename)
	default:
		log.Fatal("Invalid subcommand. Expected list, store, stat, delete, or fetch.")
	}
}
