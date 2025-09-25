package client

import (
	"file-sender/internal/protocol"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"unicode/utf8"
)

type Client struct {
	filepath   string
	serverAddr *net.TCPAddr
	meta       *protocol.FileMetadata
}

func New(filepath string, serverAddr *net.TCPAddr) (*Client, error) {
	if filepath == "" {
		return nil, fmt.Errorf("please provide a valid filepath")
	}
	filename := getFileName(filepath)
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", filename)
		}
		return nil, fmt.Errorf("error accessing file: %v", err)
	}
	fileSize := fileInfo.Size()
	if fileSize > 1<<40 {
		return nil, fmt.Errorf("file size exceeds maximum limit of 1 TB")
	}
	if utf8.RuneCountInString(filepath) > 4096 {
		return nil, fmt.Errorf("filename exceeds maximum length of 4096 bytes in UTF-8")
	}
	fileMetadata := protocol.FileMetadata{Filename: filename, FileSize: fileSize}
	return &Client{filepath: filepath, serverAddr: serverAddr, meta: &fileMetadata}, nil
}

func (cli *Client) SendFile() error {
	file, err := os.Open(cli.filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	conn, err := net.DialTCP("tcp", nil, cli.serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()
	err = protocol.WriteFileMetadata(conn, cli.meta.Filename, cli.meta.FileSize)
	if err != nil {
		return fmt.Errorf("error sending file metadata: %v", err)
	}
	bytesSent, err := io.CopyN(conn, file, cli.meta.FileSize)
	if err != nil {
		return fmt.Errorf("error sending file data: %v", err)
	}
	response := make([]byte, 1)
	_, err = conn.Read(response)
	if err != nil {
		return fmt.Errorf("error reading server response: %v", err)
	}
	if response[0] == 1 {
		fmt.Printf("File sent successfully (%d bytes)\n", bytesSent)
	} else {
		return fmt.Errorf("server failed to save the file correctly, maybe try again(?)")
	}
	return nil
}

func getFileName(path string) string {
	parsed := strings.Split(path, "/")
	return parsed[len(parsed)-1]
}
