package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type DiscordClient struct {
	c http.Client
}

func New() *DiscordClient {
	c := http.Client{Timeout: time.Duration(5) * time.Second}

	return &DiscordClient{
		c: c,
	}
}

type MessagePostBody struct {
	Content string `json:"content"`
}

func (dc *DiscordClient) WebhookSendMessage(webhookUrl string, msg *MessagePostBody) (int, string, error) {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(msgJson))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	return resp.StatusCode, string(body), nil
}
