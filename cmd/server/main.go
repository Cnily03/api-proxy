package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"api-proxy/internal/db"
	"api-proxy/internal/network"
	"api-proxy/internal/proxy"
	"api-proxy/internal/service"
)

// Entry point: parses CLI flags, initializes the SQLite database and
// services, then launches the admin panel server and reverse proxy
// server concurrently until an OS signal or fatal server error triggers
// a graceful shutdown.
func main() {
	adminPort := flag.Int("b", 8050, "admin panel port")
	proxyPort := flag.Int("p", 8060, "reverse proxy port")
	staticDir := flag.String("d", "", "serve static files from this directory")
	flag.Parse()

	dataDir := filepath.Join(".", "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("create data directory failed: %v", err)
	}

	dbPath := filepath.Join(dataDir, "data.db")
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		log.Fatalf("open sqlite failed: %v", err)
	}
	defer sqlDB.Close()

	database := db.New(sqlDB)
	if err := database.InitSchema(); err != nil {
		log.Fatalf("init schema failed: %v", err)
	}

	ruleSvc := service.NewRuleService(database)
	authSvc := service.NewAuthService(database)

	if err := authSvc.LoadConfig(); err != nil {
		log.Fatalf("load server config failed: %v", err)
	}

	adminHandler := network.Setup(ruleSvc, authSvc, *staticDir)
	proxyHandler := proxy.NewHandler(ruleSvc)

	adminServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", *adminPort),
		Handler:           adminHandler,
		ReadHeaderTimeout: 15 * time.Second,
	}

	proxyServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", *proxyPort),
		Handler:           proxyHandler,
		ReadHeaderTimeout: 30 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		log.Printf("admin panel listening on http://localhost:%d", *adminPort)
		errCh <- adminServer.ListenAndServe()
	}()
	go func() {
		log.Printf("reverse proxy listening on http://localhost:%d", *proxyPort)
		errCh <- proxyServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %v", sig)
	case srvErr := <-errCh:
		if srvErr != nil && srvErr != http.ErrServerClosed {
			log.Fatalf("server stopped with error: %v", srvErr)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown admin server error: %v", err)
	}
	if err := proxyServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown proxy server error: %v", err)
	}
}
