package main

import (
	clt "file-sender/internal/client"
	"fmt"
	"net"
	"os"
)

func main() {
	filepath := os.Args[1]
	serverHost := os.Args[2]
	serverPort := os.Args[3]
	serverAddr, err := net.ResolveTCPAddr("tcp", serverHost+":"+serverPort)
	if err != nil {
		fmt.Println("ResolveTCPAddr err:", err)
		os.Exit(1)
	}
	client, err := clt.New(filepath, serverAddr)
	if err != nil {
		fmt.Println("Error creating client:", err)
	}
	err = client.SendFile()
	if err != nil {
		fmt.Println("Error sending file:", err)
		os.Exit(1)
	}
}
