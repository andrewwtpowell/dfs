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
        stat(client, &filename)
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
func stat(client pb.DFSClient, filename *string) (*pb.MetaData, error) {

    log.Printf("Requesting statistics for file %s", filename)
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    request := pb.MetaData{
        Name: *filename,
        LockOwner: id,
    }

    fileStat, err := client.GetFileStat(ctx, &request)
    if err != nil {
        return nil, fmt.Errorf("client.GetFileStat failed: %s", err)
    }

    shared.PrintFileInfo(fileStat)

    return fileStat, nil
}

// lock attempts to lock a dfs file for editing
func lock(client pb.DFSClient, filename *string) error {

    log.Printf("Attempting to lock file %s", *filename)
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    request := pb.MetaData{
        Name: *filename,
        LockOwner: id,
    }

    response, err := client.LockFile(ctx, &request)
    if err != nil {
        return fmt.Errorf("client.LockFile failed: %s", err)
    }

    if response.LockOwner == id { 
        log.Printf("%s locked successfully")
    }

    return nil
}

// store stores a client file to the dfs server
func store(client pb.DFSClient, filename *string) {

    log.Printf("Attempting to store file %s", *filename)
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    // Get local file data
    info, err := os.Stat(*filename)
    if err != nil {
        log.Fatal(err)
    }

    // Lock file
    if err := lock(client, filename); err != nil {
        log.Fatal(err)
    }

    // Calculate client side file crc
    crc, err := shared.CalculateCrc(filename)
    if err != nil {
        log.Fatal(err)
    }

    // Determine if client has newer version of file
    serverFileStat, err := stat(client, filename) 
    if err != "File not found" {
        log.Fatal(err)
    }

    if err == nil && serverFileStat.Crc == crc {
        log.Fatalf("Client and Server have same file contents. Exiting.")
    }

    if err == nil && serverFileStat.Mtime > uint32(info.ModTime().Unix()) {
        log.Fatalf("Server has more recent version of file than client. Exiting.")
    }

    // Initialize request
    request := pb.MetaData{
        Name: *filename,
        Size: uint32(info.Size()),
        Mtime: uint32(info.ModTime().Unix()),
        Crc: crc,
        LockOwner: id,
    }

}
