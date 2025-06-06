package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	interval = 60 * 24 * time.Minute // ← 送りたい間隔を変更可
)

var startTime = time.Now()

type StateTransition struct {
	states       []string
	transitions  map[string]map[string]float64
	currentState string
	rng          *rand.Rand
}

func NewStateTransition() *StateTransition {

	elapsed := time.Since(startTime)
	seed := elapsed.Nanoseconds()

	states := []string{"うお", "🐟", "う、"}

	// 各状態からの遷移確率を定義
	transitions := map[string]map[string]float64{
		"うお": {
			"うお": 1.0,
		},
		"🐟": {
			"🐟": 1.0,
		},
		"う、": {
			"うお": 0.8,
			"う、": 0.15,
			"🐟":  0.05,
		},
	}

	return &StateTransition{
		states:       states,
		transitions:  transitions,
		currentState: states[0], // 初期状態
		rng:          rand.New(rand.NewSource(seed)),
	}
}

// 次の状態に遷移してその状態の値を返す
func (st *StateTransition) NextState() string {
	transitions := st.transitions[st.currentState]

	// 累積確率で次の状態を決定
	randVal := st.rng.Float64()
	cumulative := 0.0

	for nextState, probability := range transitions {
		cumulative += probability
		if randVal <= cumulative {
			st.currentState = nextState
			return st.currentState
		}
	}

	return st.currentState // unreachable
}

var stateTransition = NewStateTransition()

func postWebhook(ctx context.Context) error {
	webhookURL := os.Getenv("WEBHOOK_URL")

	// 確率状態遷移からpayloadを生成
	payload := stateTransition.NextState()

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

	log.Printf("sent payload: %s", payload)
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
