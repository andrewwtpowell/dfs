package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
    target = flag.String("target", "localhost:50051", "Server target to connect to")
    mountPath = flag.String("mount", "mnt/", "Server directory to mount")
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

    shared.PrintFileList(fileList)

    // Connect to server
    log.Printf("connecting to server at %s", *target)
    conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("unable to connect: %s", err)
    }
    defer conn.Close()

    //client := pb.NewDFSClient(conn)

    // Send request to server

}
