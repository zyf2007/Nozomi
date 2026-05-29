package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v3"
)

func (a *App) getResendConfig(c *gin.Context) {
	cfg, err := a.loadResendConfig(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func (a *App) putResendConfig(c *gin.Context) {
	var cfg ResendConfig
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveResendConfig(c.Param("id"), cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) saveResendConfig(providerID string, cfg ResendConfig) error {
	typeName, err := a.providerType(providerID)
	if err != nil {
		return err
	}
	if typeName != "resend" {
		return fmt.Errorf("该上游不是 Resend")
	}
	_, err = a.db.Exec(`insert into provider_resend_config(provider_id,api_key,from_address,reply_to) values(?,?,?,?) on conflict(provider_id) do update set api_key=excluded.api_key, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, cfg.APIKey, cfg.FromAddress, cfg.ReplyTo)
	return err
}

func (a *App) loadResendConfig(providerID string) (ResendConfig, error) {
	var cfg ResendConfig
	err := a.db.QueryRow(`select api_key, from_address, reply_to from provider_resend_config where provider_id=?`, providerID).Scan(&cfg.APIKey, &cfg.FromAddress, &cfg.ReplyTo)
	return cfg, err
}

func (a *App) sendResend(candidate upstreamCandidate, input MailInput) (string, error) {
	cfg := candidate.ResendConfig
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return "", fmt.Errorf("Resend API key 未配置")
	}
	from := firstNonEmpty(strings.TrimSpace(cfg.FromAddress), strings.TrimSpace(input.From))
	if from == "" {
		return "", fmt.Errorf("Resend 发信地址未配置")
	}
	if len(input.To) == 0 {
		return "", fmt.Errorf("收件人为空")
	}
	subject := firstNonEmpty(strings.TrimSpace(input.Subject), "Nozomi Relay")
	textBody := strings.TrimSpace(input.Text)
	htmlBody := strings.TrimSpace(input.HTML)
	if textBody == "" && htmlBody == "" {
		return "", fmt.Errorf("邮件正文为空")
	}

	client := resend.NewClient(apiKey)
	req := &resend.SendEmailRequest{
		From:    from,
		To:      input.To,
		Subject: subject,
	}
	if textBody != "" {
		req.Text = textBody
	}
	if htmlBody != "" {
		req.Html = htmlBody
	}
	if replyTo := strings.TrimSpace(cfg.ReplyTo); replyTo != "" {
		req.ReplyTo = replyTo
	}

	resp, err := client.Emails.SendWithContext(context.Background(), req)
	if err != nil {
		return "", err
	}
	if resp != nil && resp.Id != "" {
		return resp.Id, nil
	}
	return fmt.Sprintf("resend-%d-%d", candidate.ID, time.Now().UnixNano()), nil
}
