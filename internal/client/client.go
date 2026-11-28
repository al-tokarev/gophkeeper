// Package client реализует HTTP-транспорт к серверу GophKeeper.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Client — HTTP-клиент для API сервера GophKeeper.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New создаёт Client, указывающий на baseURL.
func New(baseURL string) (*Client, error) {
	httpClient := &http.Client{}

	caPath := os.Getenv("GOPHKEEPER_CA_CERT")
	explicit := caPath != ""
	if caPath == "" {
		for _, p := range []string{"./ca.crt", "./certs/ca.crt"} {
			if _, err := os.Stat(p); err == nil {
				caPath = p
				break
			}
		}
	}

	if caPath == "" {
		return &Client{baseURL: baseURL, http: httpClient}, nil
	}

	caCert, err := os.ReadFile(caPath)
	if err != nil {
		if explicit {
			return nil, fmt.Errorf("read CA cert %s: %w", caPath, err)
		}
		return &Client{baseURL: baseURL, http: httpClient}, nil
	}

	pool, _ := x509.SystemCertPool()
	if pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("parse CA cert %s: invalid PEM", caPath)
	}
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
	}

	return &Client{baseURL: baseURL, http: httpClient}, nil
}

// SetToken устанавливает bearer-токен для авторизованных запросов.
func (c *Client) SetToken(t string) {
	c.token = t
}

// Token возвращает текущий bearer-токен.
func (c *Client) Token() string {
	return c.token
}

// do выполняет HTTP-запрос и возвращает тело ответа.
// body == nil означает отсутствие тела запроса. auth == true добавляет bearer-токен.
// Статусы вне диапазона 2xx возвращаются как ошибка.
func (c *Client) do(ctx context.Context, method, path string, body any, auth bool) ([]byte, error) {
	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("client: marshal: %w", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("client: new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server %s: %s", resp.Status, respBody)
	}
	return respBody, nil
}
