package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
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
    if s.lockMap[metadata.GetName()] != metadata.LockOwner &&
        s.lockMap[metadata.GetName()] != "" {
        return fmt.Errorf("File %s locked by client %s", metadata.GetName(), s.lockMap[metadata.GetName()])
    }

    s.lockMap[metadata.GetName()] = metadata.GetLockOwner()
    defer func() { s.lockMap[metadata.GetName()] = "" }()

    // If file already exists on server, has been modified more recently than server verion,
    // and file data does not differ from server version, then update the MetaData for the server version
    // to match the stored file and return

    // TODO improve algorithm by sorting the fileList by file name and 
    // do binary search to find if file is present
    for _, file := range s.fileList {

        if file.Name == metadata.GetName() &&
            file.Mtime < metadata.GetMtime() &&
            file.Crc == metadata.GetCrc() {

            log.Printf("Server file contents match stored file - updating modification time for file %s", metadata.GetName())
            file.Mtime = metadata.GetMtime()

            return stream.SendAndClose(&pb.MetaData{
                Name:       file.Name,
                Size:       file.Size,
                Mtime:      file.Mtime,
                Crc:        file.Crc,
                LockOwner:  "",
            })
        }

        if file.Name == metadata.GetName() &&
            file.Mtime >= metadata.GetMtime() {

            return fmt.Errorf("File already present on server - ignoring")
        }
    }

    log.Printf("Opening file %s for writing", metadata.GetName())
    file, err := os.OpenFile(metadata.GetName(), os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        log.Print(err)
        return err
    }

    defer func() {
        if err := file.Close(); err != nil {
            log.Fatal(err)
        }
    }()

    log.Printf("Writing request contents to file %s", metadata.GetName())
    for {

        msg, err := stream.Recv()

        if err == io.EOF {
            // Reached end of stream, send file metadata response
            log.Printf("StoreFile successful for file %s", metadata.GetName())
            info, err := file.Stat()
            if err != nil {
                return err
            }
            
            _, err = shared.RefreshFileList(s.mount)
            if err != nil {
                log.Print(err)
                return err
            }

            var crc uint32
            for _, file := range s.fileList {
                if file.GetName() == info.Name() {
                    crc = file.GetCrc()
                }
            }
            
            return stream.SendAndClose(&pb.MetaData{
                Name: info.Name(),
                Size: uint32(info.Size()),
                Mtime: uint32(info.ModTime().Unix()),
                LockOwner: s.lockMap[info.Name()],
                Crc: crc,
            })
        }
        if err != nil {
            return err
        }

        // Process the request 
        data := msg.GetFiledata()
        if data == nil {
            return fmt.Errorf("Invalid chunk received in file data stream")
        }
        
        bytes, err := file.Write(data.GetContent())
        if err != nil {
            return err
        }

        log.Printf("Wrote %d bytes to file %s", bytes, metadata.GetName())
    }
}
