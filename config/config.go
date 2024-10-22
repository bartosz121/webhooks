package config

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Server  ConfigServer
	Api     ConfigApi
	Discord ConfigDiscord
}

type ConfigServer struct {
	Port         int           `env:"SERVER_PORT,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
	Debug        bool          `env:"SERVER_DEBUG,required"`
}

type ConfigApi struct {
	GpwScraperWebhookSecret           string `env:"GPW_SCRAPER_WEBHOOK_SECRET,required"`
	GpwScraperWebhookDiscordChannelId string `env:"GPW_SCRAPER_WEBHOOK_DISCORD_CHANNEL_ID,required"`
}

type ConfigDiscord struct {
	BaseUrl  string `env:"DISCORD_API_BASE_URL,required"`
	BotToken string `env:"DISCORD_BOT_TOKEN,required"`
}

func New() *Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
