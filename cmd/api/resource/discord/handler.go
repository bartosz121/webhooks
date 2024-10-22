package discord

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	errors "github.com/bartosz121/webhooks-api/cmd/api/resource/common/err"
	"github.com/bartosz121/webhooks-api/cmd/http_client/discord"
	"github.com/bartosz121/webhooks-api/config"
)

const DiscordMessageMaxLength = 2000
const BaseMessageLength = 40

type DiscordHandler struct {
	l             *zerolog.Logger
	v             *validator.Validate
	apiConfig     *config.ConfigApi
	discordClient *discord.DiscordClient
}

func NewHandler(l *zerolog.Logger, v *validator.Validate, apiConfig *config.ConfigApi, discordClient *discord.DiscordClient) *DiscordHandler {
	return &DiscordHandler{
		l:             l,
		v:             v,
		apiConfig:     apiConfig,
		discordClient: discordClient,
	}
}

type GpwScraperWebhookData struct {
	Id          int                `json:"id" validate:"required,gte=0"`
	Type        string             `json:"type" validate:"required"`
	Title       string             `json:"title" validate:"required"`
	Description *string            `json:"description"`
	Company     string             `json:"company" validate:"required"`
	Source      string             `json:"source" validate:"required"`
	ParsedByLlm *string            `json:"parsedByLlm" validate:"required"`
	Date        GpwScraperDateTime `json:"date" validate:"required"`
}

type GpwScraperDateTime struct {
	time.Time
}

func (t *GpwScraperDateTime) UnmarshalJSON(b []byte) error {
	dateString := string(b)
	dateString = dateString[1 : len(dateString)-1]
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999", dateString)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}

func createDiscordMessageContent(data *GpwScraperWebhookData) (*discord.MessagePostBody, error) {
	msg := &discord.MessagePostBody{}
	contentMaxLength := DiscordMessageMaxLength - (BaseMessageLength + len(data.Type) + len(data.Company) + len(data.Title) + len(data.Source))
	if data.Description != nil && len(*data.Description) > contentMaxLength {
		trimmedDescription := (*data.Description)[:contentMaxLength] + "..."
		// FIXME: ugh modyfing those here is dirty
		data.Type = strings.ToUpper(data.Type)
		data.Description = &trimmedDescription
	}

	tmpl := `📰 **__{{ .Type }}__**
🏢 **{{ .Company }}**
📝 **{{ .Title }}**
{{ if .Description }}` + "```" + "{{ .Description }}" + "```" + `{{ end }}{{ .Source }}`

	t, err := template.New("discordContent").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var content bytes.Buffer
	if err := t.Execute(&content, data); err != nil {
		return nil, err
	}

	msg.Content = content.String()
	return msg, nil
}

//	@Summary	Accept gpw-scraper webhook
//	@Tags		gpw-scraper
//	@Accept		json
//	@Produce	json
//	@Success	200
//	@Failure	401
//	@Failure	422
//	@Failure	424
//	@Failure	500
//	@Router		/v1/discord/gpw-scraper [post]

func (h *DiscordHandler) GpwScraperWebhook(w http.ResponseWriter, r *http.Request) {
	webhookSecretB64 := r.Header.Get("x-webhook-secret")
	h.l.Debug().Msg("x-webhook-secret header raw: " + webhookSecretB64)

	webhookSecret, err := b64.StdEncoding.DecodeString(webhookSecretB64)
	h.l.Debug().Msg("x-webhook-secret header decoded: " + string(webhookSecret) + " expected: " + h.apiConfig.GpwScraperWebhookSecret)

	if err != nil || string(webhookSecret) != h.apiConfig.GpwScraperWebhookSecret {
		errors.Unauthorized(w, []byte(`{"error": "unauthorized"}`))
		return
	}

	h.l.Debug().Str("webhook-secret", string(webhookSecret))

	webhookPayload := &GpwScraperWebhookData{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(webhookPayload); err != nil {
		h.l.Error().Msg(err.Error())
		errors.UnprocessableEntity(w, []byte(`{"error": "json error"}`))
		return
	}
	h.l.Debug().Int64("espi_ebi_id", int64(webhookPayload.Id))

	if err := h.v.Struct(webhookPayload); err != nil {
		errors.UnprocessableEntity(w, []byte(`{"error": "unexpected json data"}`))
		return
	}

	discordMsg, err := createDiscordMessageContent(webhookPayload)
	if err != nil {
		h.l.Error().Msg(err.Error())
		errors.InternalServerError(w, []byte(`{"error": "internal server error"}`))
		return
	}

	statusCode, responseBody, err := h.discordClient.SendMessage(h.apiConfig.GpwScraperWebhookDiscordChannelId, discordMsg)
	if err != nil {
		h.l.Error().Int("discord-response-status-code", statusCode).Msg(err.Error())
		errors.FailedDependency(w, []byte(`{"error": "error sending discord message"}`))
		return
	}

	if statusCode != 200 {
		h.l.Error().Int("discord-response-status-code", statusCode).Msg(responseBody)
		errors.FailedDependency(w, []byte(`{"error":`+fmt.Sprintf(`"error (%d) sending discord message"}`, statusCode)))
		return
	}

	h.l.Debug().Int("discord-response-status-code", statusCode).Msg(responseBody)

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"msg": fmt.Sprintf("webhook accepted, discord message sent (%d)", statusCode),
	}
	json.NewEncoder(w).Encode(response)
}
