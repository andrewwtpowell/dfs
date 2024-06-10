package server

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/andrewwtpowell/dfs/contract"
	"google.golang.org/grpc"
)

var (
    port = flag.Int("port", 50051, "Server port")
    mountPtr = flag.String("mount", "mnt/server", "Server directory to mount")
)

type server struct {
    pb.UnimplementedDFSServer
    mountPath string
}

func main() {

    flag.Parse()
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
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
