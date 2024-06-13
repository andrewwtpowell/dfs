package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"os"

	pb "github.com/andrewwtpowell/dfs/contract"
	"google.golang.org/grpc"
)

var (
    port = flag.Int("port", 50051, "Server port")
    mountPath = flag.String("mount", "mnt/server/", "Server directory to mount")
    fileList []pb.MetaData
)

type server struct {
    pb.UnimplementedDFSServer
}

func main() {

    flag.Parse()

    // Create directory at the mount path if it doesn't already exist
    err := os.MkdirAll(*mountPath, os.ModePerm)
    if err != nil {
        log.Fatalf("creating dir %s failed: %s", *mountPath, err)
    }

    refreshFileList()

    printFileList()

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

func refreshFileList() error {

    // Clear existing list, replaced by current mount dir contents
    fileList = nil

    files, err := os.ReadDir(*mountPath)
    if err != nil {
        log.Printf("failed to read files in %s: %s\n", *mountPath, err)
    }

    // Allocate capacity for new list based on number of files in dir
    fileList = make([]pb.MetaData, 0, len(files))

    for _, file := range files {

        // Skip subdirectories - future functionality
        if file.IsDir() {
            continue
        }

        filePath := *mountPath + file.Name()

        // Initialize MetaData object and add to list
        var newFileEntry pb.MetaData
        newFileEntry.Name = file.Name()

        fileInfo, err := file.Info()
        if err != nil {
            log.Fatalf("unable to get info for %s: %s", filePath, err)
        }

        newFileEntry.Size = uint32(fileInfo.Size())
        newFileEntry.Mtime = uint32(fileInfo.ModTime().Unix())

        openedFile, err := os.Open(filePath)
        if err != nil {
            log.Fatalf("failed to open file %s: %s\n", filePath, err)
        }
        defer openedFile.Close()

        fileContents, err := os.ReadFile(filePath)
        if err != nil {
            log.Fatalf("unable to read contents of %s: %s\n", filePath, err)
        }

        newFileEntry.Crc = crc32.Checksum(fileContents, crc32.IEEETable)

        fileList = append(fileList, newFileEntry)
    }

    return nil
}

func printFileList() {

    for _, data := range fileList {
        fmt.Printf("%v\n", data.Name)
        fmt.Printf("size:\t%v\n", data.Size)
        fmt.Printf("mtime:\t%v\n", data.Mtime)
        fmt.Printf("crc:\t%v\n", data.Crc)
    }

}
