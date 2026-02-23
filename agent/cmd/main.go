package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/querywise/agent/internal/client"
	"github.com/querywise/agent/internal/collector"
	"github.com/querywise/agent/internal/config"
)

func main() {
	configPath := flag.String("config", "config/agent.yml", "Path to configuration file")
	flag.Parse()

	log.Printf("QueryWise Agent starting...")
	log.Printf("Config file: %s", *configPath)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.SSLMode)

	c, err := collector.New(dsn, cfg.Agent.Name, cfg.Collection.BatchSize, cfg.Collection.TimeoutSeconds)
	if err != nil {
		log.Fatalf("Collector: %v", err)
	}
	defer c.Close()

	ingestClient := client.New(cfg.Backend.URL, cfg.Backend.APIKey)
	interval := time.Duration(cfg.Collection.IntervalSeconds) * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first collection soon
	go func() {
		time.Sleep(2 * time.Second)
		runCollection(ctx, c, ingestClient)
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Agent shutting down")
			return
		case <-ticker.C:
			runCollection(ctx, c, ingestClient)
		}
	}
}

func runCollection(ctx context.Context, c *collector.Collector, client *client.Client) {
	payload, err := c.Collect(ctx)
	if err != nil {
		log.Printf("Collect error: %v", err)
		return
	}
	if len(payload.Queries) == 0 {
		log.Printf("No query metrics collected (pg_stat_statements may be empty or extension disabled)")
		return
	}
	if err := client.Ingest(payload); err != nil {
		log.Printf("Ingest error: %v", err)
		return
	}
	log.Printf("Ingested %d queries for db_id=%s", len(payload.Queries), payload.DBID)
}
