package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const deadlineTimeout = 7

var (
    server = os.Getenv("DFS_SERVER_ADDR")
    fileList []pb.MetaData
    id string
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
    id = name + fmt.Sprint(pid)
    log.Printf("starting client %s", id)

    // Refresh client file list
    clientDir, err := os.Getwd()
    if err != nil {
        log.Fatalf("os.Getwd failed: %s", err)
    }

    fileList, err := shared.RefreshFileList(clientDir + "/mnt/")
    if err != nil {
        log.Fatalf("RefreshFileList: %s", err)
    }

    shared.PrintFileList(&fileList)

    // Connect to server
    log.Printf("connecting to server at %s", server)
    conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("unable to connect: %s", err)
    }
    defer conn.Close()
    client := pb.NewDFSClient(conn)

    if len(os.Args) < 2 {
        log.Fatal("Expected subcommand. Subcommand options are list, store, stat, delete, or fetch.")
    }

    switch os.Args[1] {
    case "list":
        list(client)
    case "store":
    case "stat":
        filename := os.Args[2]
        if filename == "" {
            log.Fatalf("No file name provided for stat RPC")
        }
        log.Printf("Calling stat RPC with file name %s", filename)
        stat(client, filename)
    case "delete":
    case "fetch":
    default:
        log.Fatal("Invalid subcommand. Expected list, store, stat, delete, or fetch.")
    }
}

// list gets files present on the server and prints them to stdout
func list(client pb.DFSClient) {

    log.Printf("Requesting list of server-side files")
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    request := pb.MetaData{ Name: "" }
    list, err := client.ListFiles(ctx, &request)
    if err != nil {
        log.Fatalf("client.ListFiles failed: %s", err)
    }

    log.Printf("Received list containing %d files", len(list.FileList))
    for _, file := range list.FileList {
        log.Printf("%s", file.GetName())
    }
}

// stat gets file statistics for a server-side file
func stat(client pb.DFSClient, filename string) {

    log.Printf("Requesting statistics for file %s", filename)
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    request := pb.MetaData{
        Name: filename,
        LockOwner: id,
    }

    fileStat, err := client.GetFileStat(ctx, &request)
    if err != nil {
        log.Fatalf("client.GetFileStat failed: %s", err)
    }

    shared.PrintFileInfo(fileStat)
}
