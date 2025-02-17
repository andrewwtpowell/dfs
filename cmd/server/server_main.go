package main

import (
    "flag"
    "log"
	"net"
    "fmt"
	"google.golang.org/grpc"

    "github.com/andrewwtpowell/dfs/api"
    "github.com/andrewwtpowell/dfs/pkg/server"
)

var (
	port          = flag.Int("port", 50051, "Server port")
	mountPath     = flag.String("mount", "mnt/", "Server directory to mount")
)

func main() {

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	dfs_api.RegisterDFSServer(grpcServer, server.NewServer(*mountPath))

	log.Printf("server listening at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
