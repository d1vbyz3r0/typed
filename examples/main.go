package main

import (
	"examples/server"
	"log"
	"os"
)

func main() {
	os.Mkdir("uploads", 0755)
	srv := server.NewBuilder().Build()
	log.Fatal(srv.Start(":8080"))
}
