package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	interval = 60 * 24 * time.Minute // ← 送りたい間隔を変更可
	payload  = "うお"                  // POST する本文
)

func postWebhook(ctx context.Context) error {
	webhookURL := os.Getenv("WEBHOOK_URL")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func main() {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// すぐ 1 回目を送る
	if err := postWebhook(context.Background()); err != nil {
		log.Printf("initial post failed: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := postWebhook(context.Background()); err != nil {
				log.Printf("post failed: %v", err)
			} else {
				log.Printf("posted successfully at %s", time.Now().Format(time.RFC3339))
			}
		}
	}
}
