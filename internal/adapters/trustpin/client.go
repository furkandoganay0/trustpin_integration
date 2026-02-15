package trustpin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
	retry   RetryConfig
}

type RetryConfig struct {
	Max       int
	Backoff   time.Duration
}

type Error struct {
	Status int
	Body   []byte
}

func (e *Error) Error() string {
	return "trustpin_error"
}

func NewClient(baseURL, apiKey string, timeout time.Duration, retry RetryConfig) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: timeout},
		retry:   retry,
	}
}

func (c *Client) do(ctx context.Context, method, path, tenantID string, payload any, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := c.baseURL + path
	for attempt := 0; attempt <= c.retry.Max; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("X-Tenant-ID", tenantID)

		resp, err := c.client.Do(req)
		if err != nil {
			if attempt < c.retry.Max {
				time.Sleep(c.retry.Backoff)
				continue
			}
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil {
				return nil
			}
			if err := json.Unmarshal(body, out); err != nil {
				return err
			}
			return nil
		}

		if resp.StatusCode == http.StatusServiceUnavailable && attempt < c.retry.Max {
			time.Sleep(c.retry.Backoff)
			continue
		}

		return &Error{Status: resp.StatusCode, Body: body}
	}

	return errors.New("retry_exhausted")
}
