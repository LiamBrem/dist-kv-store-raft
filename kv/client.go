package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a simple HTTP client for the KV store
type Client struct {
	baseURL string
}

// NewClient creates a new KV store client
func NewClient(host string, port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
	}
}

// Put stores a key-value pair
func (c *Client) Put(key, value string) error {
	reqBody, err := json.Marshal(map[string]string{
		"key":   key,
		"value": value,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/put", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PUT failed: %s", string(body))
	}

	return nil
}

// Get retrieves a value for a key
func (c *Client) Get(key string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/get?key=%s", c.baseURL, key))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GET failed: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["value"], nil
}

// GetAllKeys retrieves all keys in the store
func (c *Client) GetAllKeys() ([]string, error) {
	resp, err := http.Get(c.baseURL + "/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result["keys"], nil
}
