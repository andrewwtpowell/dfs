package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/andrewwtpowell/dfs/contract"
	"google.golang.org/grpc"
    "github.com/andrewwtpowell/dfs/shared"
)

var (
    port = flag.Int("port", 50051, "Server port")
    mountPath = flag.String("mount", "mnt/", "Server directory to mount")
    fileList []pb.MetaData
)

type server struct {
    pb.UnimplementedDFSServer
}

func main() {

    flag.Parse()

    fileList, err := shared.RefreshFileList(*mountPath)
    if err != nil {
        log.Fatalf("refreshFileList: %s", err)
    }

    shared.PrintFileList(fileList)

    startServer(*port)
}

func startServer(port int) {

    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    dfsServer := grpc.NewServer()
    pb.RegisterDFSServer(dfsServer, &server{})

    log.Printf("server listening at %v", listener.Addr())
    if err := dfsServer.Serve(listener); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
