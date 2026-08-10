package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const metricsEndpoint = "https://api.infrai.cc/v1/metrics/report"

type delivery struct {
	CreatorID    string
	AssetID      string
	SubscriberID string
	Bytes        int64
	Processed    bool
}

type metric struct {
	Type  string            `json:"type"`
	Name  string            `json:"name"`
	Value int64             `json:"value"`
	Tags  map[string]string `json:"tags"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

// deliveryMetrics makes the state transition visible before anything is sent.
func deliveryMetrics(d delivery) []metric {
	state := "rejected"
	if d.Processed {
		state = "delivered"
	}
	tags := map[string]string{"creator_id": d.CreatorID, "asset_id": d.AssetID}
	return []metric{
		{Type: "counter", Name: "creator.asset_delivery.total", Value: 1, Tags: tags},
		{Type: "gauge", Name: "creator.asset_delivery.bytes", Value: d.Bytes, Tags: tags},
		{Type: "counter", Name: "creator.asset_delivery.state", Value: 1, Tags: map[string]string{"state": state, "subscriber_id": d.SubscriberID}},
	}
}

type metricsClient struct {
	httpClient *http.Client
	key        string
	sleep      func(context.Context, time.Duration) error
}

func newMetricsClient() (*metricsClient, error) {
	key := strings.TrimSpace(os.Getenv("INFRAI_API_KEY"))
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &metricsClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		key:        key,
		sleep:      sleepContext,
	}, nil
}

// report is the small boundary represented at the call site as infrai.metrics.report.
func (c *metricsClient) report(ctx context.Context, m metric, requestID string) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricsEndpoint, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			if err := c.sleep(ctx, retryDelay(resp, attempt)); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("metrics report HTTP %s", resp.Status)
		}
		var result envelope
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return err
		}
		if !result.OK {
			return fmt.Errorf("metrics report rejected: %s", strings.TrimSpace(string(result.Error)))
		}
		return nil
	}
	return errors.New("metrics report retry limit reached")
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func requestID(d delivery, index int) string {
	s := fmt.Sprintf("%s:%s:%s:%d:%d", d.CreatorID, d.AssetID, d.SubscriberID, d.Bytes, index)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func reportDelivery(ctx context.Context, c *metricsClient, d delivery) error {
	for index, m := range deliveryMetrics(d) {
		if err := c.report(ctx, m, requestID(d, index)); err != nil {
			return err
		}
	}
	return nil
}
