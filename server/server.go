package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
    port = flag.Int("port", 50051, "Server port")
    mountPath = flag.String("mount", "mnt/", "Server directory to mount")
)

type dfsServer struct {
    pb.UnimplementedDFSServer
    fileList    []pb.MetaData
    mount       string
    queueMutex  sync.Mutex
    dataMutex   sync.Mutex
    lockMap     map[string]string
}

func main() {

    flag.Parse()

    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterDFSServer(grpcServer, newServer())

    log.Printf("server listening at %v", listener.Addr())
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

func newServer() *dfsServer {
    s := &dfsServer{mount: *mountPath}
    list, err := shared.RefreshFileList(s.mount)
    if err != nil {
        log.Fatalf("refreshFileLIst: %s", err)
    }

    log.Printf("Found files at %s: \n", s.mount)
    s.fileList = list
    shared.PrintFileList(&s.fileList)

    return s
}

func (s *dfsServer) StoreFile(stream pb.DFS_StoreFileServer) error {

    log.Println("Processing store request")

    s.dataMutex.Lock()
    defer s.dataMutex.Unlock()

    // Implement based on c++ implementation and go route_guide example

    // Process first message in stream (metadata)
    msg, err := stream.Recv()
    if err != nil {
        return err
    }

    metadata := msg.GetMetadata()
    if metadata == nil {
        return fmt.Errorf("Initial request message must contain MetaData")
    }

    // If file is locked return error, otherwise lock the file
    // defer file unlock

    // If file already exists on server, has been modified more recently than server verion,
    // and file data does not differ from server version, then update the MetaData for the server version
    // to match the stored file and return

    // Store new/updated file

    for {

        msg, err := stream.Recv()

        if err == io.EOF {
            // Reached end of stream, send file metadata response
            return stream.SendAndClose(&pb.MetaData{})
        }
        if err != nil {
            return err
        }

        // Process the request 
        
        
        
    }

    // Refresh server file list
}
