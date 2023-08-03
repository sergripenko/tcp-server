// Package main - contains main function to run client.
package main

import (
	"fmt"
	"log"

	"tcp-server/internal/client"
	"tcp-server/internal/config"
)

func main() {
	conf, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("%s:%d", conf.ServerHost, conf.ServerPort)
	cl := client.NewClient(addr, conf.HashcashMaxIterations)

	log.Println("start client")

	if err = cl.Run(); err != nil {
		log.Fatal(err)
	}
}
