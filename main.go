package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"homestead/internal/checker"
	"homestead/internal/config"
	"homestead/internal/server"
)

//go:embed web
var webFS embed.FS

func main() {
	configPath := flag.String("config", "config.toml", "Path to TOML config file")
	host := flag.String("host", "127.0.0.1", "Host/IP to bind (use :: for all IPv6, 0.0.0.0 for all IPv4)")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Environment variables override flags.
	if v := os.Getenv("HOMESTEAD_CONFIG"); v != "" {
		*configPath = v
	}
	if v := os.Getenv("HOMESTEAD_HOST"); v != "" {
		*host = v
	}
	if v := os.Getenv("HOMESTEAD_PORT"); v != "" {
		*port = v
	}
	// Legacy PORT env (common in container environments).
	if v := os.Getenv("PORT"); v != "" && os.Getenv("HOMESTEAD_PORT") == "" {
		*port = v
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("Loaded config: %q (%d sections)", cfg.Title, len(cfg.Sections))

	chk := checker.New(cfg)
	chk.Start()
	defer chk.Stop()

	srv := server.New(cfg, *configPath, chk, *host, *port, webFS)

	// Graceful shutdown on SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down…")
		os.Exit(0)
	}()

	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
