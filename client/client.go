package main

import (
	"context"
	"flag"
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
    server = flag.String("server", "localhost:50051", "Server to connect to in the format ip:port")
    mountPath = flag.String("mount", "mnt/", "Server directory to mount")
    rpc = flag.String("rpc", "list", "rpc to call (list, store, stat, delete, fetch)")
    fileList []pb.MetaData
    id string
)

func main() {

    flag.Parse()

    name, err := os.Hostname()
    if err != nil {
        log.Fatalf("os.Hostname: %s", err)
    }
    pid := os.Getpid()
    id = name + fmt.Sprint(pid)
    log.Printf("starting client %s", id)

    // Get files at mount point
    fileList, err := shared.RefreshFileList(*mountPath)
    if err != nil {
        log.Fatalf("RefreshFileList: %s", err)
    }

    shared.PrintFileList(&fileList)

    // Connect to server
    log.Printf("connecting to server at %s", *server)
    conn, err := grpc.NewClient(*server, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("unable to connect: %s", err)
    }
    defer conn.Close()

    //client := pb.NewDFSClient(conn)

    // Send request to server

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

    log.Printf("%+v", *fileStat)

}
