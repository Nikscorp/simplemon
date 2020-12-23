package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	opts, err := parseOpts()
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("[INFO] Opts parsed successfully")
	config, err := parseConfig(opts.Config)
	if err != nil {
		log.Fatalf("[CRITICAL] Failed to parse config: %v", err)
	}
	log.Printf("[INFO] Config parsed successfully")
	log.Printf("[DEBUG] %+v", config)
	sch := NewScheduler(*config)
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("[WARN] Got signal %v. Shutting down", sig)
		cancel()
	}()
	sch.Run(ctx)
	sch.Done()
	log.Printf("[WARN] Done")
}
