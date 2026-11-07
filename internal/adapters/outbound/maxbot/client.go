package maxbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

func New(botToken, chatID string) *Client {
	return &Client{
		botToken:   botToken,
		chatID:     chatID,
		httpClient: &http.Client{Timeout: 10},
	}
}

type sendMessageRequest struct {
	Text string `json:"text"`
}

func (c *Client) SendText(ctx context.Context, text string) error {
	reqBody, err := json.Marshal(sendMessageRequest{
		Text: text,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	u := fmt.Sprintf("https://platform-api2.max.ru/messages?chat_id=%s", url.QueryEscape(c.chatID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MAX Bot API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
