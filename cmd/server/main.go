// Package main - contains main function to run server.
package main

import (
	"log"

	"tcp-server/internal/config"
	"tcp-server/internal/server"
)

func main() {
	conf, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("start server")

	serv := server.NewServer(conf)

	if err = serv.Run(); err != nil {
		log.Fatal(err)
	}
}
