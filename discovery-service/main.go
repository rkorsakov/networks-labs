package main

import (
	"discovery-service/discoveryservice"
	"fmt"
	"os"
	"time"
)

func main() {
	ds, err := discoveryservice.NewDiscoveryService(os.Args[1], 5*time.Second)
	if err != nil {
		fmt.Printf("Error creating discoveryservice: %s", err)
		os.Exit(1)
	}
	ds.Start()
	select {}
}
