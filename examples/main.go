package main

import (
	"log"
	"os"

	"github.com/d1vbyz3r0/typed/examples/server"
)

func main() {
	os.Mkdir("uploads", 0755)
	srv := server.NewBuilder().Build()
	log.Fatal(srv.Start(":8080"))
}
