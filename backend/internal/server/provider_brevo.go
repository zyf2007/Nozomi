package server

import (
	"context"
	"fmt"
	stdhtml "html"
	"strings"
	"time"

	brevo "github.com/getbrevo/brevo-go/lib"
	"github.com/gin-gonic/gin"
)

func (a *App) getBrevoConfig(c *gin.Context) {
	cfg, err := a.loadBrevoConfig(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func (a *App) putBrevoConfig(c *gin.Context) {
	var cfg BrevoConfig
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveBrevoConfig(c.Param("id"), cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) saveBrevoConfig(providerID string, cfg BrevoConfig) error {
	typeName, err := a.providerType(providerID)
	if err != nil {
		return err
	}
	if typeName != "brevo" {
		return fmt.Errorf("该上游不是 Brevo")
	}
	_, err = a.db.Exec(`insert into provider_brevo_config(provider_id,api_key,from_address,from_name,reply_to) values(?,?,?,?,?) on conflict(provider_id) do update set api_key=excluded.api_key, from_address=excluded.from_address, from_name=excluded.from_name, reply_to=excluded.reply_to`, providerID, cfg.APIKey, cfg.FromAddress, firstNonEmpty(cfg.FromName, "Nozomi"), cfg.ReplyTo)
	return err
}

func (a *App) loadBrevoConfig(providerID string) (BrevoConfig, error) {
	var cfg BrevoConfig
	err := a.db.QueryRow(`select api_key, from_address, from_name, reply_to from provider_brevo_config where provider_id=?`, providerID).Scan(&cfg.APIKey, &cfg.FromAddress, &cfg.FromName, &cfg.ReplyTo)
	return cfg, err
}

func (a *App) sendBrevo(candidate upstreamCandidate, input MailInput) (string, error) {
	cfg := candidate.BrevoConfig
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return "", fmt.Errorf("Brevo API key 未配置")
	}

	from := firstNonEmpty(strings.TrimSpace(cfg.FromAddress), strings.TrimSpace(input.From))
	if from == "" {
		return "", fmt.Errorf("Brevo 发信地址未配置")
	}
	subject := firstNonEmpty(strings.TrimSpace(input.Subject), "Nozomi Relay")
	textBody := strings.TrimSpace(input.Text)
	htmlBody := strings.TrimSpace(input.HTML)
	if len(input.To) == 0 {
		return "", fmt.Errorf("收件人为空")
	}
	if htmlBody == "" {
		if textBody == "" {
			return "", fmt.Errorf("邮件正文为空")
		}
		htmlBody = "<pre style=\"white-space:pre-wrap;word-break:break-word;\">" + stdhtml.EscapeString(textBody) + "</pre>"
	}

	apiClient := brevo.NewAPIClient(brevo.NewConfiguration())
	ctx := context.WithValue(context.Background(), brevo.ContextAPIKey, brevo.APIKey{Key: apiKey})
	req := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Name:  firstNonEmpty(strings.TrimSpace(cfg.FromName), "Nozomi"),
			Email: from,
		},
		To: []brevo.SendSmtpEmailTo{
			{Email: input.To[0]},
		},
		Subject:     subject,
		HtmlContent: htmlBody,
		TextContent: textBody,
	}
	if replyTo := strings.TrimSpace(cfg.ReplyTo); replyTo != "" {
		req.ReplyTo = &brevo.SendSmtpEmailReplyTo{Email: replyTo}
	}
	if len(input.To) > 1 {
		req.To = make([]brevo.SendSmtpEmailTo, 0, len(input.To))
		for _, rcpt := range input.To {
			req.To = append(req.To, brevo.SendSmtpEmailTo{Email: rcpt})
		}
	}

	resp, _, err := apiClient.TransactionalEmailsApi.SendTransacEmail(ctx, req)
	if err != nil {
		return "", err
	}
	if resp.MessageId != "" {
		return resp.MessageId, nil
	}
	return fmt.Sprintf("brevo-%d-%d", candidate.ID, time.Now().UnixNano()), nil
}
