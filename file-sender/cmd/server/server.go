package main

import (
	"context"
	srvr "file-sender/internal/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	serverPort := os.Args[1]
	server := srvr.NewServer(serverPort)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		fmt.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Stop(shutdownCtx); err != nil {
			fmt.Printf("server shutdown error: %v", err)
		}
		cancel()
	}()
	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}

}
