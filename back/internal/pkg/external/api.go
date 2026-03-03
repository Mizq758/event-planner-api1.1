package external

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    httpClient *http.Client
    apiURL     string
}

type QuoteResponse struct {
    Content string `json:"content"`
    Author  string `json:"author"`
}

func NewClient(apiURL string, timeout int) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: time.Duration(timeout) * time.Second,
        },
        apiURL: apiURL,
    }
}

// GetRandomQuote получает случайную цитату
func (c *Client) GetRandomQuote(ctx context.Context) (string, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to call external API: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("external API returned status: %d", resp.StatusCode)
    }

    var quote QuoteResponse
    if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return fmt.Sprintf("✨ Idea: \"%s\" — %s", quote.Content, quote.Author), nil
}