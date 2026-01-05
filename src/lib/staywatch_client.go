package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StayWatchClient はStayWatch APIへのリクエストを管理するクライアント
type StayWatchClient struct {
	apiKey string
	client *http.Client
}

// NewStayWatchClient は新しいStayWatchClientを作成する
func NewStayWatchClient(apiKey string) *StayWatchClient {
	return &StayWatchClient{
		apiKey: apiKey,
		client: SharedHTTPClient,
	}
}

// Get はGETリクエストを送信し、結果をresultにデコードする
func (c *StayWatchClient) Get(url string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
