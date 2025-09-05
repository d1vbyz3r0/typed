package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/d1vbyz3r0/typed/internal/generator"
)

var (
	configPath = flag.String("config", "", "path to config file")
	version    = flag.Bool("version", false, "print version and exit")
)

func GetVersion() (version string) {
	if b, ok := debug.ReadBuildInfo(); ok && len(b.Main.Version) > 0 {
		version = b.Main.Version
	} else {
		version = "development"
	}
	return
}

func main() {
	flag.Parse()

	if *version {
		fmt.Println(GetVersion())
		return
	}

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
