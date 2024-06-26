package main

import (
	"flag"
	abciserver "github.com/cometbft/cometbft/abci/server"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	db "kvstore/database"
)

var homeDir string
var socketAddr string

func init() {
	flag.StringVar(&homeDir, "home", "", "Path to the kvstore directory (if empty, uses $HOME/.kvstore)")
	flag.StringVar(&socketAddr, "address", "unix://example.sock", "Unix domain socket address (if empty, uses \"unix://example.sock\"")
}

func main() {
	flag.Parse()

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))

	defaultHomeDir := "$HOME/.kvstore++"
	if homeDir == "" {
		homeDir = os.ExpandEnv(defaultHomeDir)
	}
	dbPath := filepath.Join(homeDir, "data")
	db, err := db.NewPebbleDB("kvstore++", dbPath)
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	logger.Info("database start", "folder", dbPath)

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Closing database: %v", err)
		}
	}()

	app := NewKVStoreApplication(db, logger)

	server := abciserver.NewSocketServer(socketAddr, app)
	server.SetLogger(logger)

	if err := server.Start(); err != nil {
		logger.Error("server", "error starting socket server", err)
		os.Exit(1)
	}
	if server.IsRunning() {
		logger.Info("service running", "msg", "ABCI Server started")
	}

	defer server.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
