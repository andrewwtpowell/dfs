package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/andrewwtpowell/dfs/contract"
	"github.com/andrewwtpowell/dfs/shared"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const deadlineTimeout = 7

var (
	server    = os.Getenv("DFS_SERVER_ADDR")
	fileList  []*pb.MetaData
	id        string
	fileMutex sync.Mutex
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
	case "mount":
		mountPath := os.Args[2]
		if mountPath == "" {
			mountPath = "mnt/"
		}
		fileList = getFileList(&mountPath)
		mount(client, &mountPath)
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
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for delete RPC")
		}
		deleteFile(client, &filename)
	case "fetch":
		filename := os.Args[2]
		if filename == "" {
			log.Fatalf("No file name provided for fetch RPC")
		}
		fetch(client, &filename)
	default:
		log.Fatal("Invalid subcommand. Expected list, store, stat, delete, or fetch.")
	}
}

func mount(client pb.DFSClient, mountPath *string) {

	if err := os.Chdir(*mountPath); err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Updated application working directory: ", wd)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Filter out files that aren't .png / .jpeg / .txt / etc.
				if !(strings.Contains(event.Name, ".png") ||
					strings.Contains(event.Name, ".jpeg") ||
					strings.Contains(event.Name, ".txt") ||
					strings.Contains(event.Name, ".html")) {
					continue
				}

                if event.Has(fsnotify.Create) {
                    continue
                }

                fileMutex.Lock()

				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file: ", event.Name)
					store(client, &event.Name)
				}
				if event.Has(fsnotify.Remove) {
					log.Println("removed file: ", event.Name)
					deleteFile(client, &event.Name)
				}

                fileMutex.Unlock()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error: ", err)
			}
		}
	}()

	err = watcher.Add(wd)
	if err != nil {
		log.Fatal(err)
	}

	serverSync(client, mountPath)
}

func getFileList(mountDir *string) []*pb.MetaData {

	log.Printf("updating file list with files at %s", *mountDir)
	fileList, err := shared.RefreshFileList(mountDir)
	if err != nil {
		log.Fatalf("RefreshFileList: %s", err)
	}

	shared.PrintFileList(&fileList)

	return fileList
}

// list gets files present on the server and prints them to stdout
func list(client pb.DFSClient) {

	log.Print("Requesting list of server-side files")
	ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
	defer cancel()

	request := pb.MetaData{Name: ""}
	list, err := client.ListFiles(ctx, &request)
	if err != nil {
		log.Fatalf("client.ListFiles failed: %s", err)
	}

	log.Printf("Received list containing %d files:", len(list.FileList))
	for _, file := range list.FileList {
		fmt.Printf("%+v\n", file)
	}
}

// deleteFile removes file from server
func deleteFile(client pb.DFSClient, filepath *string) {

	log.Printf("Attempting to delete file %s", *filepath)
	filename := getFilenameFromPath(*filepath)
	log.Printf("Filename: %s", filename)

	_, err := stat(client, &filename)
	st := status.Convert(err)
	if st.Code() != codes.OK {
		log.Fatal(st.String())
	}

	err = lock(client, &filename)
	st = status.Convert(err)
	if st.Code() != codes.OK {
		log.Fatal(st.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
	defer cancel()

	request := pb.MetaData{
		Name:      filename,
		LockOwner: id,
	}

	_, err = client.DeleteFile(ctx, &request)
	st = status.Convert(err)
	if st.Code() != codes.OK {
		log.Fatal(st.String())
	}

	log.Printf("File %s deleted successfully", filename)
}

// stat gets file statistics for a server-side file
func stat(client pb.DFSClient, filename *string) (*pb.MetaData, error) {

	log.Printf("Requesting statistics for file %s", *filename)
	ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
	defer cancel()

	request := pb.MetaData{
		Name:      *filename,
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
		Name:      *filename,
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

func getFilenameFromPath(filepath string) string {
	return filepath[strings.LastIndex(filepath, "/")+1:]
}

// serverSync continuously sends requests for server updates and receives server updates
func serverSync(client pb.DFSClient, mountPath *string) {

	log.Printf("Starting server sync RPC")
	ctx := context.WithoutCancel(context.Background())

	stream, err := client.ServerSync(ctx)
	st := status.Convert(err)
	if st.Code() != codes.OK {
		log.Fatal(st.String())
	}

	for {

		// Create empty request
		request := pb.MetaData{
			Name:      "",
			Size:      0,
			Mtime:     0,
			Crc:       0,
			LockOwner: id,
		}

		// Send request to receive server update
		stream.Send(&request)

		response, err := stream.Recv()
		if err != nil {
			log.Fatalf("serverSync stream recv failed")
		}

		serverList := response.GetFileList()

		// Delete local files that are not in server list
		localToDelete := findMissing(serverList, fileList)
		for _, toDelete := range localToDelete {
			log.Printf("Deleting file %s locally, not on server", *mountPath+toDelete.Name)
			os.Remove(*mountPath + toDelete.Name)

			// Remove entry from local list in place
			i := 0
			for i < len(fileList) {
				if fileList[i].Name == toDelete.Name {
					copy(fileList[i:], fileList[i+1:])
					fileList = fileList[:len(fileList)-1]
				} else {
					i++
				}
			}
		}

		// Fetch files from server that are not present locally
		missingLocally := findMissing(fileList, serverList)
		for _, toFetch := range missingLocally {
			log.Printf("Fetching file %s from server, missing locally", toFetch.Name)
			fetch(client, &toFetch.Name)
		}

		// Store more recent local files to server - push updates
		updatedLocal := findUpdated(serverList, fileList)
		for _, toStore := range updatedLocal {
			log.Printf("Storing file %s, client has more recent version than server", toStore.Name)
			store(client, &toStore.Name)
		}

		// Fetch more recent server files from server - pull updates
		updatedServer := findUpdated(fileList, serverList)
		for _, toFetch := range updatedServer {
			log.Printf("Fetching file %s from server, client has stale version", toFetch.Name)
			fetch(client, &toFetch.Name)
		}
	}
}

// findUpdated returns the elements in 'compare' that are more recent compared to their equivalent in 'base'
func findUpdated(base, compare []*pb.MetaData) []*pb.MetaData {

	var updated []*pb.MetaData
	for _, compareItem := range compare {
		for _, baseItem := range base {
			if compareItem.Name == baseItem.Name && compareItem.Mtime > baseItem.Mtime {
				updated = append(updated, compareItem)
			}
		}
	}
	return updated
}

// findMissing returns the elements in 'compare' that are not in 'base'
func findMissing(base, compare []*pb.MetaData) []*pb.MetaData {

	var missing []*pb.MetaData
	for _, compareItem := range compare {
		found := false
		for _, baseItem := range base {
			if compareItem.Name == baseItem.Name {
				found = true
				break
			}
		}
		if !found {
			log.Printf("File %s missing from base list", compareItem.Name)
			missing = append(missing, compareItem)
		}
	}

	return missing
}

// store stores a client file to the dfs server
func store(client pb.DFSClient, filepath *string) {

	log.Printf("Attempting to store file %s", *filepath)
	filename := getFilenameFromPath(*filepath)
	log.Printf("Filename: %s", filename)
	ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
	defer cancel()

	// Get local file data
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Calculate client side file crc
	crc, err := shared.CalculateCrc(&filename)
	if err != nil {
		log.Fatal(err)
	}

	// Determine if client has newer version of file
	serverFileStat, err := stat(client, &filename)
	st := status.Convert(err)
	if st.Code() == codes.NotFound {
		log.Printf("Caught file not found error: %s", st.String())
	} else if st.Code() != codes.OK {
		log.Fatal(st.String())
	}

	if err == nil && serverFileStat.Crc == crc {
		log.Printf("Client and Server already have same file contents, no need to store.")
		return
	}

	if err == nil && serverFileStat.Mtime > int32(info.ModTime().Unix()) {
		log.Printf("Server has more recent version of file than client, aborting store operation.")
		return
	}

	// Lock file
	if err := lock(client, &filename); err != nil {
		log.Fatal(err)
	}

	// Initialize request
	metadata := pb.MetaData{
		Name:      filename,
		Size:      int32(info.Size()),
		Mtime:     int32(info.ModTime().Unix()),
		Crc:       crc,
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
	log.Printf("Data buffer size: %d bytes", bufSize)

	file, err := os.Open(*filepath)
	defer file.Close()
	if err != nil {
		log.Fatalf("os.Open failed: %s", err)
	}
	buf := make([]byte, bufSize)

	for {

		numBytes, err := file.Read(buf)
		if err == io.EOF && numBytes == 0 {
			break
		}

		request := pb.StoreRequest_Content{Content: buf}
		msg := &pb.StoreRequest{RequestData: &request}

		if (err == io.EOF && numBytes != 0) || len(buf) == 0 {
			log.Printf("reached end of file, sending final %d bytes", len(buf))
			if err := stream.Send(msg); err != nil {
				log.Fatalf("client.StoreFile: stream.Send(%v) failed: %s", *msg, err)
			}
			break
		}

		if err != nil {
			log.Fatalf("file.Read failed: %s", err)
		}

		//log.Printf("sending %d bytes", len(buf))
		if err := stream.Send(msg); err != nil {
			log.Fatalf("client.StoreFile: stream.Send(%v) failed: %s", *msg, err)
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

// fetch fetches a file from the server and stores it on the client
func fetch(client pb.DFSClient, filename *string) {

	log.Printf("Requesting file %s from server", *filename)

	ctx, cancel := context.WithTimeout(context.Background(), deadlineTimeout*time.Second)
	defer cancel()

    fileMutex.Lock()
    defer fileMutex.Unlock()

	request := pb.MetaData{
		Name:      *filename,
		LockOwner: id,
	}
	stream, err := client.FetchFile(ctx, &request)
	if err != nil {
		log.Fatalf("client.FetchFile failed: %s", err)
	}

	msg, err := stream.Recv()
	if err != nil {
		log.Fatalf("stream.Recv failed: %s", err)
	}

	metadata := msg.GetMetadata()
	if metadata == nil {
		log.Fatal("Received incorrect response type from server")
	}

	if metadata.Crc <= 0 {
		log.Fatal("Server file CRC not provided")
	}

	file, err := os.OpenFile(*filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	defer file.Close()
	if err != nil && os.IsExist(err) {
		crc, err := shared.CalculateCrc(filename)
		if err != nil {
			log.Fatalf("CalculateCrc failed: %s", err)
		}

		if crc == metadata.Crc {
			log.Printf("Client already has up to date version of file %s, exiting", *filename)
			return
		}

		file, err = os.OpenFile(*filename, os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("os.OpenFile failed: %s", err)
		}
		defer file.Close()

	} else if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error opening file %s: %s", *filename, err)
	}

	log.Printf("Writing response contents to file %s", *filename)

	for {

		msg, err := stream.Recv()

		if err == io.EOF {

			log.Printf("Fetch successful for file %s", *filename)

			crc, err := shared.CalculateCrc(filename)
			if err != nil {
				log.Fatalf("CalculateCrc failed: %s", err)
			}

			if crc != metadata.Crc {
				log.Printf("CRC mismatch. Server CRC: %d, Client CRC: %d", metadata.Crc, crc)
			}

			return
		}
		if err != nil {
			log.Fatalf("stream.Recv failed: %s", err)
		}

		data := msg.GetContent()
		if data == nil {
			log.Fatalf("invalid chunk received in file data stream")
		}

		_, err = file.Write(data)
		if err != nil {
			log.Fatalf("writing content to file failed")
		}
	}
}
