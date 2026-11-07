package maxbot

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

func New(botToken, chatID string) *Client {
	// Загружаем CA-бандл из SSL_CERT_FILE или MAX_CA_CERT_PATH
	caPaths := []string{
		os.Getenv("MAX_CA_CERT_PATH"),
		os.Getenv("SSL_CERT_FILE"),
	}

	var caCertPool *x509.CertPool
	for _, path := range caPaths {
		if path == "" {
			continue
		}
		caCert, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		break
	}

	// Если нет CA-файла — пробуем системный (работает на Linux)
	if caCertPool == nil {
		roots, _ := x509.SystemCertPool()
		if roots != nil {
			caCertPool = roots
		}
	}

	transport := &http.Transport{
		// Отключаем HTTP/2 — сервер MAX может некорректно обрабатывать
		TLSNextProto: make(map[string]func(*http.Client, string, *tls.Conn)),
		TLSClientConfig: &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
		},
		// Настройки подключения
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		botToken: botToken,
		chatID:   chatID,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
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
	// Принудительно HTTP/1.1
	req.Close = true

	slog.Info("sending to MAX", slog.String("url", u), slog.String("text", text))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("MAX response", slog.Int("status", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MAX Bot API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
