package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bartosz121/webhooks-api/config"
)

type DiscordClient struct {
	c    http.Client
	conf *config.ConfigDiscord
}

func New(conf *config.ConfigDiscord) *DiscordClient {
	c := http.Client{Timeout: time.Duration(5) * time.Second}

	return &DiscordClient{
		c:    c,
		conf: conf,
	}
}

type MessagePostBody struct {
	Content string `json:"content"`
}

func (dc *DiscordClient) SendMessage(channelId string, msg *MessagePostBody) (int, string, error) {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/channels/%s/messages", dc.conf.BaseUrl, channelId), bytes.NewBuffer(msgJson))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", dc.conf.BotToken))
	req.Header.Set("User-Agent", "Discord Bot (https://webhooks.bartoszmagiera.dev, 0.0.1)")

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
