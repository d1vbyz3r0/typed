package main

import (
	"flag"
	"github.com/d1vbyz3r0/typed/internal/generator"
	"log"
	"log/slog"
	"os"
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

	if cfg.Debug {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		})
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}

	g, err := generator.New(cfg)
	if err != nil {
		log.Fatalf("create generator: %v", err)
	}

	err = g.Generate()
	if err != nil {
		log.Fatalf("failed to generate spec builder: %v", err)
	}
}
