package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
        filename := os.Args[2]
        if filename == "" {
            log.Fatal("No file name provided for store RPC")
        }
        store(client, &filename)
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

    log.Print("Requesting list of server-side files")
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

    log.Printf("Requesting statistics for file %s", *filename)
    ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
    defer cancel()

    request := pb.MetaData{
        Name: *filename,
        LockOwner: id,
    }

    fileStat, err := client.GetFileStat(ctx, &request)
    st := status.Convert(err)
    if st.Code() != codes.OK {
        log.Print(st.String())
        return nil, st.Err()
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
    st := status.Convert(err)
    if st.Code() != codes.OK {
        log.Print(st.String())
        return st.Err()
    }

    if response.LockOwner == id { 
        log.Printf("%s locked successfully", *filename)
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
    st := status.Convert(err)
    if st.Code() == codes.NotFound {
        log.Printf("Caught file not found error: %s", st.String())
    } 
    if st.Code() != codes.OK {
        log.Fatal(st.String())
    }

    if err == nil && serverFileStat.Crc == crc {
        log.Fatalf("Client and Server have same file contents. Exiting.")
    }

    if err == nil && serverFileStat.Mtime > int32(info.ModTime().Unix()) {
        log.Fatalf("Server has more recent version of file than client. Exiting.")
    }

    // Initialize request
    metadata := pb.MetaData{
        Name: *filename,
        Size: int32(info.Size()),
        Mtime: int32(info.ModTime().Unix()),
        Crc: crc,
        LockOwner: id,
    }
    request := pb.StoreRequest_Metadata{
        Metadata: &metadata,
    }
    msg := &pb.StoreRequest{RequestData: &request}

    stream, err := client.StoreFile(ctx)
    st = status.Convert(err)
    if st.Code() != codes.OK {
        log.Fatal(st.String())
    }

    // Send metadata as inital request in stream
    stream.Send(msg)

    var bufSize int
    if int(info.Size()) > shared.MaxBufSize {
        bufSize = shared.MaxBufSize
    } else {
        bufSize = int(info.Size())
    }

    file, err := os.Open(*filename)
    defer file.Close()
    if err != nil {
        log.Fatalf("os.Open failed: %s", err) 
    }
    buf := make([]byte, bufSize)

    for {

        numBytes, readErr := file.Read(buf)
        if readErr != nil && readErr != io.EOF {
            log.Fatalf("file.Read failed: %s", readErr)
        }

        filedata := pb.FileData {
            Content: buf,
            Size: int32(numBytes),
        }
        request := pb.StoreRequest_Filedata{
            Filedata: &filedata,
        }
        msg := &pb.StoreRequest{RequestData: &request}
        if err := stream.Send(msg); err != nil {
            log.Fatalf("client.StoreFile: stream.Send(%v) failed: %s", msg, err)
        }

        if readErr == io.EOF {
            break
        } 
    }

    response, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatalf("StoreFile failed: %s", err)
    }

    log.Print("Received StoreFile response from server: ")
    shared.PrintFileInfo(response)

    if crc != response.Crc {
        log.Fatalf("Client/Server CRC mismatch. Client CRC: %d, Server CRC: %d", crc, response.Crc)
    }
}
