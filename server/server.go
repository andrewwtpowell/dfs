package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

var (
    port = flag.Int("port", 50051, "Server port")
    mountPath = flag.String("mount", "mnt/", "Server directory to mount")
)

type dfsServer struct {
    pb.UnimplementedDFSServer
    fileList    []*pb.MetaData
    mount       string
    queueMutex  sync.Mutex
    dataMutex   sync.Mutex
    lockMap     map[string]string
}

func main() {

    flag.Parse()

    listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
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
        log.Fatalf("refreshFileList: %s", err)
    }

    log.Printf("Found files at %s: \n", s.mount)
    s.fileList = list
    shared.PrintFileList(&s.fileList)

    s.lockMap = make(map[string]string)

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
        errMsg := fmt.Sprintf("stream.Recv failed: %s", err)
        log.Print(err)
        err := status.Error(codes.Canceled, errMsg)
        return err
    }

    metadata := msg.GetMetadata()
    if metadata == nil {
        errMsg := fmt.Sprint("initial message in stream must contain MetaData")
        log.Print(errMsg)
        err := status.Error(codes.InvalidArgument, errMsg)
        return err
    }

    // If file is locked return error, otherwise lock the file
    if s.lockMap[metadata.Name] != metadata.LockOwner &&
        s.lockMap[metadata.Name] != "" {
        errMsg := fmt.Sprintf("file %s locked by client %s", metadata.Name, metadata.LockOwner)
        log.Printf(errMsg)
        err := status.Errorf(codes.ResourceExhausted, errMsg)
        return err
    }

    s.lockMap[metadata.Name] = metadata.LockOwner
    defer func() { s.lockMap[metadata.Name] = "" }()

    // If file already exists on server, has been modified more recently than server verion,
    // and file data does not differ from server version, then update the MetaData for the server version
    // to match the stored file and return

    // TODO improve algorithm by sorting the fileList by file name and 
    // do binary search to find if file is present
    for _, file := range s.fileList {

        if file.Name == metadata.Name &&
            file.Mtime < metadata.Mtime &&
            file.Crc == metadata.Crc {

            log.Printf("Server file contents match stored file - updating modification time for file %s", metadata.Name)
            file.Mtime = metadata.Mtime

            return stream.SendAndClose(&pb.MetaData{
                Name:       file.Name,
                Size:       file.Size,
                Mtime:      file.Mtime,
                Crc:        file.Crc,
                LockOwner:  "",
            })
        }

        if file.Name == metadata.Name &&
            file.Mtime >= metadata.Mtime {

            log.Printf("File %s already present on server - ignoring", metadata.Name)
            return nil
        }
    }

    mountedFileLoc := s.mount + metadata.Name
    log.Printf("Opening file %s for writing", mountedFileLoc)
    file, err := os.OpenFile(mountedFileLoc, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        errMsg := fmt.Sprintf("error opening file %s: %s", mountedFileLoc, err)
        log.Print(errMsg)
        err := status.Errorf(codes.Canceled, errMsg)
        return err
    }

    defer func() {
        if err := file.Close(); err != nil {
            log.Fatalf("error closing file %s: %s", mountedFileLoc, err)
        }
    }()

    log.Printf("Writing request contents to file %s", mountedFileLoc)
    for {

        msg, err := stream.Recv()

        if err == io.EOF {
            // Reached end of stream, send file metadata response
            log.Printf("StoreFile successful for file %s", mountedFileLoc)
            info, err := file.Stat()
            if err != nil {
                errMsg := fmt.Sprintf("Stat failed for stored file: %s", err)
                log.Print(errMsg)
                err := status.Errorf(codes.Canceled, errMsg)
                return err
            }
            
            _, err = shared.RefreshFileList(s.mount)
            if err != nil {
                errMsg := fmt.Sprintf("shared.RefreshFileList failed: %s", err)
                log.Print(errMsg)
                err := status.Errorf(codes.Canceled, errMsg)
                return err
            }

            var crc uint32
            for _, file := range s.fileList {
                if file.Name == info.Name() {
                    crc = file.Crc
                }
            }
            
            return stream.SendAndClose(&pb.MetaData{
                Name: info.Name(),
                Size: int32(info.Size()),
                Mtime: int32(info.ModTime().Unix()),
                LockOwner: s.lockMap[info.Name()],
                Crc: crc,
            })
        }
        if err != nil {
            errMsg := fmt.Sprintf("stream.Recv failed: %s", err)
            log.Print(errMsg)
            err := status.Errorf(codes.Unknown, errMsg)
            return err
        }

        // Process the request 
        data := msg.GetFiledata()
        if data == nil {
            errMsg := fmt.Sprintf("invalid chunk received in file data stream")
            log.Print(errMsg)
            err := status.Error(codes.InvalidArgument, errMsg)
            return err
        }
        
        bytes, err := file.Write(data.GetContent())
        if err != nil {
            errMsg := fmt.Sprint("writing content to file failed")
            log.Print(errMsg)
            err := status.Error(codes.Canceled, errMsg)
            return err
        }

        log.Printf("Wrote %d bytes to file %s", bytes, mountedFileLoc)
    }
}

func (s *dfsServer) ListFiles(ctx context.Context, request *pb.MetaData) (*pb.ListResponse, error) {

    log.Printf("Received request for file list")

    s.dataMutex.Lock()
    defer s.dataMutex.Unlock()

    log.Printf("File list contains %d files", len(s.fileList))
    var response pb.ListResponse
    response.FileList = s.fileList

    log.Printf("Response list contains %d files", len(response.FileList))
    return &response, nil
}

func (s *dfsServer) DeleteFile(ctx context.Context, request *pb.MetaData) (*pb.MetaData, error) {

    s.dataMutex.Lock()
    defer s.dataMutex.Unlock()

    defer shared.RefreshFileList(s.mount)

    // If file is locked return error, otherwise lock the file
    if s.lockMap[request.Name] != request.LockOwner &&
        s.lockMap[request.Name] != "" {

        errMsg := fmt.Sprintf("File %s locked by client %s", request.Name, s.lockMap[request.Name])
        log.Print(errMsg)
        err := status.Error(codes.ResourceExhausted, errMsg)
        return nil, err
    }

    // Check server list for file
    for _, file := range s.fileList {

        if(request.Name == file.Name) {
            os.Remove(request.Name)
            delete(s.lockMap, request.Name)
            return file, nil
        }
    }

    errMsg := fmt.Sprintf("File %s not found", request.Name)
    log.Print(errMsg)
    err := status.Error(codes.NotFound, errMsg)
    return nil, err
}

func (s *dfsServer) GetFileStat(ctx context.Context, request *pb.MetaData) (*pb.MetaData, error) {

    log.Printf("Client %s requesting info for file %s", request.LockOwner, request.Name)
    shared.PrintFileList(&s.fileList)
    s.dataMutex.Lock()
    defer s.dataMutex.Unlock()

    for _, file := range s.fileList {

        if file.Name == request.Name {
            log.Printf("Found file %s", request.Name)
            response := &pb.MetaData{
                Name:       file.Name,
                Size:       file.Size,
                Mtime:      file.Mtime,
                LockOwner:  s.lockMap[file.Name],
                Crc:        file.Crc,
            }
            shared.PrintFileInfo(response)

            return response, nil
        }
    }

    errMsg := fmt.Sprintf("File %s not found", request.Name)
    log.Print(errMsg)
    err := status.Error(codes.NotFound, errMsg)
    return nil, err
}

func (s *dfsServer) LockFile(ctx context.Context, request *pb.MetaData) (*pb.MetaData, error) {

    log.Printf("Client %s attempting to lock file %s", request.LockOwner, request.Name)
    
    if request.LockOwner == "" {
        err := status.Error(codes.InvalidArgument, "No client ID supplied in lock request")
        return nil, err
    }

    if s.lockMap[request.Name] == "" {
        s.lockMap[request.Name] = request.LockOwner
        log.Printf("Client %s locked file %s", request.LockOwner, request.Name)
        return request, nil
    } 

    if s.lockMap[request.Name] == request.LockOwner {
        log.Printf("Client %s already owns lock for file %s", request.LockOwner, request.Name)
        return request, nil
    } 

    err := status.Errorf(codes.ResourceExhausted, "File %s not available. Locked by client %s", request.Name, s.lockMap[request.Name])
    return nil, err
}
