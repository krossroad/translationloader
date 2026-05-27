package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/app"
)

var (
	localesFlag = flag.String("locales", "", "Comma-separated list of locales: Defaults to 'en'")
	prodIDArgs  = flag.String("products", "", "Comma-separated list of product-ids")
)

func main() {
	flag.Parse()
	if *prodIDArgs == "" {
		fmt.Println("Usage: sync --products id1,id2")
		os.Exit(1)
	}
	productIDs := strings.Split(*prodIDArgs, ",")

	cfg := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncApp, err := app.NewSyncApplication(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer syncApp.Close()

	docs, err := syncApp.BuildProductDocument(ctx, productIDs)
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	for _, doc := range docs {
		fmt.Printf("Assembled Document for ID: %s\n%+v\n\n", doc.UUID, doc)
	}
}

// loadConfig
func loadConfig() app.AppConfig {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	locales := []string{"en"}
	if *localesFlag != "" {
		locales = strings.Split(*localesFlag, ",")
	}

	ttlStr := os.Getenv("CACHE_TTL")
	ttl, _ := time.ParseDuration(ttlStr)
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	capacityStr := os.Getenv("CACHE_OTTER_CAPACITY")
	capacity, _ := strconv.Atoi(capacityStr)
	if capacity == 0 {
		capacity = 1000
	}

	return app.AppConfig{
		DBDSN: dsn,
		Cache: cache.Config{
			Driver:   os.Getenv("CACHE_DRIVER"),
			TTL:      ttl,
			Capacity: capacity,
		},
		Locales: locales,
	}
}
