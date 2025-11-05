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
)

var targetDate = time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC)

func getDaysUntilTarget() int {
	now := time.Now().UTC()
	// 日付のみで比較するため、時刻を0時に設定
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	diff := targetDate.Sub(today)
	days := int(diff.Hours() / 24)
	return days
}

func postWebhook(ctx context.Context) error {
	webhookURL := os.Getenv("WEBHOOK_URL")
	
	days := getDaysUntilTarget()
	var payload string
	if days > 0 {
		payload = fmt.Sprintf("試験日まであと%d日", days)
	} else if days == 1 {
		payload = "覚悟"
	} else {
		payload = fmt.Sprintf("2025年11月16日から%d日経過しました", -days)
	}
	
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
