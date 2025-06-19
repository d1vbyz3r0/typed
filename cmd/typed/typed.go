package main

import (
	"flag"
	"github.com/d1vbyz3r0/typed/generator"
	"log"
)

var (
	configPath = flag.String("config", "", "path to config file")
)

func main() {
	flag.Parse()

	if *configPath == "" {
		log.Fatal("config path not provided")
	}

	cfg, err := generator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	g := generator.New(cfg)
	if err := g.Generate(); err != nil {
		log.Fatalf("Failed to generate spec builder: %v", err)
	}
}
