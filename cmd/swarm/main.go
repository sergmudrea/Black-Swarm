// Black Swarm 3.0 (Siege) — Distributed scanning system with genetic optimisation.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blackswarm/siege/internal/config"
	"github.com/blackswarm/siege/internal/crypto"
	"github.com/blackswarm/siege/internal/node"
	"github.com/blackswarm/siege/internal/utils"
)

func main() {
	mode := flag.String("mode", "", "node mode: strategic, scanner, hybrid")
	configPath := flag.String("config", "", "path to configuration file")
	encryptIn := flag.String("encrypt-in", "", "path to plaintext JSON config to encrypt")
	encryptOut := flag.String("encrypt-out", "", "path to write encrypted config")
	flag.Parse()

	// Encrypt config mode
	if *encryptIn != "" || *encryptOut != "" {
		if *encryptIn == "" || *encryptOut == "" {
			fmt.Fprintf(os.Stderr, "both -encrypt-in and -encrypt-out are required\n")
			os.Exit(1)
		}
		keyEnv := os.Getenv("SIEGE_CONFIG_KEY")
		if keyEnv == "" {
			fmt.Fprintf(os.Stderr, "SIEGE_CONFIG_KEY environment variable not set\n")
			os.Exit(1)
		}
		key := []byte(keyEnv)
		if len(key) != 32 {
			fmt.Fprintf(os.Stderr, "SIEGE_CONFIG_KEY must be 32 bytes\n")
			os.Exit(1)
		}
		plain, err := os.ReadFile(*encryptIn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read input file: %v\n", err)
			os.Exit(1)
		}
		ciphertext, err := crypto.Encrypt(plain, key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "encryption failed: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*encryptOut, ciphertext, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write encrypted config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Config encrypted successfully.")
		return
	}

	// Normal run modes
	if *mode != "strategic" && *mode != "scanner" && *mode != "hybrid" {
		fmt.Fprintf(os.Stderr, "usage: swarm -mode=strategic|scanner|hybrid -config=<path>\n")
		os.Exit(1)
	}
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "config path is required\n")
		os.Exit(1)
	}

	logger := utils.NewLogger(slog.LevelInfo, os.Stdout)

	keyEnv := os.Getenv("SIEGE_CONFIG_KEY")
	if keyEnv == "" {
		logger.Error("SIEGE_CONFIG_KEY environment variable not set")
		os.Exit(1)
	}
	key := []byte(keyEnv)
	if len(key) != 32 {
		logger.Error("SIEGE_CONFIG_KEY must be 32 bytes")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configPath, key)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Override mode from flag if set
	if *mode != "" {
		cfg.Mode = *mode
	}

	swarmNode := node.NewNode(cfg, logger)
	if err := swarmNode.Start(); err != nil {
		logger.Error("failed to start node", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("received signal, shutting down", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := swarmNode.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}
}
