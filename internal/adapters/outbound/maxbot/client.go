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
			slog.Warn("failed to read CA cert", slog.String("path", path), slog.String("error", err.Error()))
			continue
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		slog.Info("loaded CA certs", slog.String("path", path))
		break
	}

	// Если нет CA-файла — пробуем системный (работает на Linux)
	if caCertPool == nil {
		roots, err := x509.SystemCertPool()
		if err != nil {
			slog.Warn("failed to load system cert pool", slog.String("error", err.Error()))
		}
		if roots != nil {
			caCertPool = roots
			slog.Info("using system cert pool")
		}
	}

	// TLS с явным HTTP/1.1 — сервер MAX не поддерживает HTTP/2
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"http/1.1"},
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		DisableKeepAlives: true,
		MaxIdleConns:      0,
		IdleConnTimeout:   30 * time.Second,
	}

	return &Client{
		botToken: botToken,
		chatID:   chatID,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   45 * time.Second,
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

	slog.Info("sending to MAX", slog.String("url", u), slog.String("text", text))

	// Создаём НОВЫЙ контекст с таймаутом для HTTP-запроса
	reqCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, u, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("MAX response received", slog.Int("status", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MAX Bot API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
