package server

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *App) getSMTPConfig(c *gin.Context) {
	cfg, err := a.loadSMTPConfig(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func (a *App) putSMTPConfig(c *gin.Context) {
	var cfg SMTPConfig
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveSMTPConfig(c.Param("id"), cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) saveSMTPConfig(providerID string, cfg SMTPConfig) error {
	typeName, err := a.providerType(providerID)
	if err != nil {
		return err
	}
	if typeName != "smtp" {
		return fmt.Errorf("该上游不是 SMTP")
	}
	if strings.TrimSpace(cfg.Password) == "" {
		_, err = a.db.Exec(`insert into provider_smtp_config(provider_id,host,port,username,password,from_address,reply_to) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set host=excluded.host, port=excluded.port, username=excluded.username, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, cfg.Host, cfg.Port, cfg.Username, "", cfg.FromAddress, cfg.ReplyTo)
		return err
	}
	_, err = a.db.Exec(`insert into provider_smtp_config(provider_id,host,port,username,password,from_address,reply_to) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set host=excluded.host, port=excluded.port, username=excluded.username, password=excluded.password, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.FromAddress, cfg.ReplyTo)
	return err
}

func (a *App) loadSMTPConfig(providerID string) (SMTPConfig, error) {
	var cfg SMTPConfig
	err := a.db.QueryRow(`select host, port, username, password, from_address, reply_to from provider_smtp_config where provider_id=?`, providerID).Scan(&cfg.Host, &cfg.Port, &cfg.Username, &cfg.Password, &cfg.FromAddress, &cfg.ReplyTo)
	return cfg, err
}

func (a *App) sendSMTPUpstream(candidate upstreamCandidate, input MailInput) (string, string, error) {
	cfg := candidate.SMTPConfig
	if strings.TrimSpace(cfg.Host) == "" {
		return "", "", fmt.Errorf("SMTP 上游主机未配置")
	}
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	from := firstNonEmpty(strings.TrimSpace(cfg.FromAddress), input.From)
	raw := buildMimeMessage(from, input.To, input.Subject, input.Text, input.HTML, input.Headers, input.Attachments)
	var auth smtp.Auth
	if strings.TrimSpace(cfg.Username) != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	if err := smtp.SendMail(hostPort, auth, from, input.To, []byte(raw)); err != nil {
		return "", raw, err
	}
	return fmt.Sprintf("smtp-%d-%d", candidate.ID, time.Now().UnixNano()), raw, nil
}
