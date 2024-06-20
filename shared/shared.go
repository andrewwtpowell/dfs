package shared

import (
    "fmt"
    "os"
	"hash/crc32"
    "log"

	pb "github.com/andrewwtpowell/dfs/contract"
)

func RefreshFileList(mountDir string) ([]pb.MetaData, error) {

    err := os.MkdirAll(mountDir, os.ModePerm)
    if err != nil {
        log.Fatalf("creating dir %s failed: %s", mountDir, err)
    }

    // Clear existing list, replaced by current mount dir contents
    var fileList []pb.MetaData

    files, err := os.ReadDir(mountDir)
    if err != nil {
        log.Printf("failed to read files in %s: %s\n", mountDir, err)
    }

    // Allocate capacity for new list based on number of files in dir
    fileList = make([]pb.MetaData, 0, len(files))

    for _, file := range files {

        // Skip subdirectories - future functionality
        if file.IsDir() {
            continue
        }

        filePath := mountDir + file.Name()

        // Initialize MetaData object and add to list
        var newFileEntry pb.MetaData
        newFileEntry.Name = file.Name()

        fileInfo, err := file.Info()
        if err != nil {
            return nil, fmt.Errorf("unable to get info for directory %s: %s", mountDir, err)
        }

        newFileEntry.Size = uint32(fileInfo.Size())
        newFileEntry.Mtime = uint32(fileInfo.ModTime().Unix())

        openedFile, err := os.Open(filePath)
        if err != nil {
            return nil, fmt.Errorf("failed to open file %s: %s\n", filePath, err)
        }
        defer openedFile.Close()

        fileContents, err := os.ReadFile(filePath)
        if err != nil {
            return nil, fmt.Errorf("unable to read contents of %s: %s\n", filePath, err)
        }

        newFileEntry.Crc = crc32.Checksum(fileContents, crc32.IEEETable)

        fileList = append(fileList, newFileEntry)
    }

    return fileList, nil
}

func PrintFileList(list *[]pb.MetaData) {

    log.Print("mounted files include:")
    for _, data := range *list {
        fmt.Printf("%v\n", data.Name)
        fmt.Printf("size:\t%v\n", data.Size)
        fmt.Printf("mtime:\t%v\n", data.Mtime)
        fmt.Printf("crc:\t%v\n", data.Crc)
    }

}
